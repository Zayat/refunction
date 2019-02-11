package worker

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"os"
	"runtime"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	. "github.com/ostenbom/refunction/worker/state"
)

func NewWorker(id string, client *containerd.Client, runtime, targetSnapshot string) (*Worker, error) {
	ctx := namespaces.WithNamespace(context.Background(), "refunction-worker"+id)

	snapManager, err := NewSnapshotManager(ctx, client, runtime)
	if err != nil {
		return nil, err
	}

	return &Worker{
		ID:             id,
		targetSnapshot: targetSnapshot,
		runtime:        runtime,
		client:         client,
		ctx:            ctx,
		creator:        cio.NullIO,
		snapManager:    snapManager,
	}, nil
}

type Worker struct {
	ID             string
	ContainerID    string
	targetSnapshot string
	runtime        string
	client         *containerd.Client
	ctx            context.Context
	creator        cio.Creator
	snapManager    *SnapshotManager
	container      containerd.Container
	task           containerd.Task
	taskExitChan   <-chan containerd.ExitStatus
	attached       bool
	stopped        bool
	IP             net.IP
}

func (m *Worker) WithCreator(creator cio.Creator) {
	m.creator = creator
}

func WithNetNsHook(ipFile string) oci.SpecOpts {
	return func(_ context.Context, _ oci.Client, _ *containers.Container, s *oci.Spec) error {
		s.Hooks = &specs.Hooks{
			Prestart: []specs.Hook{specs.Hook{
				Path: "/usr/local/bin/netns",
				Args: []string{"netns", "--ipfile", ipFile},
			}},
		}
		return nil
	}
}

func (m *Worker) Start() error {
	err := m.snapManager.CreateLayerFromBase(m.targetSnapshot)
	if err != nil {
		return err
	}

	m.ContainerID = fmt.Sprintf("%s-%s-%d", m.targetSnapshot, m.ID, rand.Intn(100))
	_, err = m.snapManager.GetRwMounts(m.targetSnapshot, m.ContainerID)
	if err != nil {
		return err
	}

	var processArgs []string
	if m.runtime != "alpine" {
		processArgs = []string{m.runtime, m.targetSnapshot}
	} else {
		processArgs = []string{m.targetSnapshot}
	}

	ipFile, err := ioutil.TempFile("", "container-ip")
	ipFileName := ipFile.Name()
	if err != nil {
		return fmt.Errorf("could not make tmp ip file: %s", err)
	}
	err = ipFile.Close()
	if err != nil {
		return fmt.Errorf("could not close container ip file: %s", err)
	}

	container, err := m.client.NewContainer(
		m.ctx,
		m.ContainerID,
		containerd.WithSnapshot(m.ContainerID),
		containerd.WithNewSpec(WithNetNsHook(ipFileName), oci.WithProcessArgs(processArgs...)),
	)
	if err != nil {
		return fmt.Errorf("could not create worker container: %s", err)
	}

	m.container = container

	task, err := container.NewTask(m.ctx, m.creator)
	if err != nil {
		return fmt.Errorf("could not create worker task: %s", err)
	}
	m.task = task

	taskExitChan, err := task.Wait(m.ctx)
	if err != nil {
		return fmt.Errorf("could not create worker task channel: %s", err)
	}
	m.taskExitChan = taskExitChan

	err = task.Start(m.ctx)
	if err != nil {
		return fmt.Errorf("could not start worker task: %s", err)
	}

	ipBytes, err := ioutil.ReadFile(ipFileName)
	if err != nil {
		return fmt.Errorf("could not read container ip file: %s", err)
	}

	m.IP = net.ParseIP(string(ipBytes))

	m.attached = false
	m.stopped = false

	return nil
}

func (m *Worker) AwaitOnline() error {
	tcpAddr := net.TCPAddr{
		IP:   m.IP,
		Port: 5000,
	}

	fmt.Println(m.IP)
	fmt.Printf("Before dial: %d\n", time.Now().UnixNano())
	dialStart := time.Now()
	conn, err := net.DialTCP("tcp", nil, &tcpAddr)
	if err != nil {
		return fmt.Errorf("could not dial worker: %s", err)
	}
	defer conn.Close()
	fmt.Printf("dial time: %s\n", time.Since(dialStart))

	writeBytes := []byte("hello there!\n")
	_, err = conn.Write(writeBytes)
	if err != nil {
		return fmt.Errorf("could not write to worker: %s", err)
	}

	result, err := ioutil.ReadAll(conn)
	if err != nil {
		return fmt.Errorf("could not read from worker: %s", err)
	}
	if len(result) != len(writeBytes) {
		return fmt.Errorf("worker did not echo")
	}

	return nil
}

func (m *Worker) Attach() error {
	// Crucial: trying to detach from a different thread
	// than the attacher causes undefined behaviour
	runtime.LockOSThread()
	err := syscall.PtraceAttach(int(m.task.Pid()))
	if err != nil {
		return err
	}

	m.attached = true
	m.stopped = true

	_, err = syscall.Wait4(int(m.task.Pid()), nil, 0, nil)
	return err
}

func (m *Worker) Detach() error {
	if !m.stopped {
		err := m.Stop()
		if err != nil {
			return fmt.Errorf("could not stop child for detach: %s", err)
		}
	}

	err := syscall.PtraceDetach(int(m.task.Pid()))
	if err != nil {
		return err
	}

	m.attached = false
	m.stopped = false
	runtime.UnlockOSThread()
	return nil
}

func (m *Worker) Stop() error {
	err := syscall.Kill(int(m.task.Pid()), syscall.SIGSTOP)
	if err != nil {
		return err
	}

	_, err = syscall.Wait4(int(m.task.Pid()), nil, 0, nil)
	m.stopped = true
	return err
}

func (m *Worker) Continue() error {
	m.stopped = false
	return syscall.PtraceCont(int(m.task.Pid()), 0)
}

func (m *Worker) SendEnableSignal() error {
	pid := int(m.task.Pid())

	err := syscall.Kill(pid, syscall.SIGUSR1)
	if err != nil {
		return err
	}

	// If not attached, signal will go through
	if !m.attached {
		return nil
	}

	var waitStat syscall.WaitStatus
	_, err = syscall.Wait4(int(pid), &waitStat, 0, nil)
	if err != nil {
		return err
	}
	if !waitStat.Stopped() {
		return errors.New("child not stopped after signal")
	}

	return syscall.PtraceCont(int(pid), int(waitStat.StopSignal()))
}

func (m *Worker) GetState() (*State, error) {
	if !m.stopped {
		err := m.Stop()
		if err != nil {
			return nil, fmt.Errorf("could not stop child to get state: %s", err)
		}
		defer m.Continue()
	}

	state, err := NewState(int(m.task.Pid()))
	if err != nil {
		return nil, fmt.Errorf("could not get state: %s", err)
	}

	return state, nil
}

func (m *Worker) SetRegs(state *State) error {
	if !m.stopped {
		err := m.Stop()
		if err != nil {
			return fmt.Errorf("could not stop child to set regs: %s", err)
		}
		defer m.Continue()
	}

	err := state.RestoreRegs()
	if err != nil {
		return fmt.Errorf("could not set regs: %s", err)
	}

	return nil
}

func (m *Worker) ClearMemRefs() error {
	pid := int(m.task.Pid())
	f, err := os.OpenFile(fmt.Sprintf("/proc/%d/clear_refs", pid), os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("could not open clear_refs for pid %d: %s", pid, err)
	}
	defer f.Close()

	_, err = f.WriteString("4")
	if err != nil {
		return fmt.Errorf("could not clear_refs for pid %d: %s", pid, err)
	}
	return nil
}

func (m *Worker) GetImage(name string) (containerd.Image, error) {
	return m.client.GetImage(m.ctx, name)
}

func (m *Worker) ListImages() ([]containerd.Image, error) {
	return m.client.ListImages(m.ctx)
}

func (m *Worker) Pid() (uint32, error) {
	if m.task == nil {
		return 0, errors.New("child not initialized")
	}

	return m.task.Pid(), nil
}

func (m *Worker) PrintStack() error {
	return m.printProcFile("stack")
}

func (m *Worker) PrintMaps() error {
	return m.printProcFile("maps")
}

func (m *Worker) printProcFile(fileName string) error {
	stack, err := ioutil.ReadFile(fmt.Sprintf("/proc/%d/%s", int(m.task.Pid()), fileName))
	if err != nil {
		return fmt.Errorf("could not read child /proc/pid/%s file: %s", fileName, err)
	}

	fmt.Println(string(stack))
	return nil
}

func (m *Worker) PrintRegs() error {
	err := m.Stop()
	if err != nil {
		return fmt.Errorf("could not stop child for regs print: %s", err)
	}
	defer m.Continue()

	var regs syscall.PtraceRegs
	err = syscall.PtraceGetRegs(int(m.task.Pid()), &regs)
	if err != nil {
		return fmt.Errorf("could not get regs: %s", err)
	}

	fmt.Printf("Regs: %+v\n", regs)

	return nil
}

func (m *Worker) End() error {
	if m.task != nil {
		if err := m.task.Kill(m.ctx, syscall.SIGKILL); err != nil {
			if errdefs.IsFailedPrecondition(err) || errdefs.IsNotFound(err) {
				return nil
			}
			return fmt.Errorf("failed to kill in manager end: %s", err)
		}

		<-m.taskExitChan

		_, err := m.task.Delete(m.ctx)
		if err != nil {
			return err
		}
	}

	if m.container != nil {
		m.container.Delete(m.ctx, containerd.WithSnapshotCleanup)
	}

	return nil
}

func (m *Worker) CleanSnapshot(name string) error {
	sservice := m.client.SnapshotService("overlayfs")
	return sservice.Remove(m.ctx, name)
}
