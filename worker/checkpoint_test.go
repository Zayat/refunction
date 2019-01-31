package worker_test

import (
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/ostenbom/refunction/worker"
)

var _ = Describe("Worker Manager checkpointing", func() {
	var id string

	BeforeEach(func() {
		id = strconv.Itoa(GinkgoParallelNode())
	})

	Describe("memory checkpointing", func() {
		var worker *Worker
		var targetLayer string

		JustBeforeEach(func() {
			var err error
			worker, err = NewWorker(id, client, targetLayer)
			Expect(err).NotTo(HaveOccurred())

			Expect(worker.Start()).To(Succeed())
		})

		AfterEach(func() {
			err := worker.End()
			Expect(err).NotTo(HaveOccurred())
		})

		Context("for loop stack + heap", func() {
			BeforeEach(func() {
				targetLayer = "forloopheap"
			})

			It("can clear memory refs", func() {
				Expect(worker.Attach()).To(Succeed())
				Expect(worker.ClearMemRefs()).To(Succeed())

				state, err := worker.GetState()
				Expect(err).NotTo(HaveOccurred())

				// while still stopped, we expect there to be no dirty pages
				dirtyStack, err := state.CountDirtyPages("[stack]")
				dirtyHeap, err2 := state.CountDirtyPages("[heap]")
				Expect(worker.Detach()).To(Succeed())

				Expect(err).NotTo(HaveOccurred())
				Expect(dirtyStack).To(Equal(0))
				Expect(err2).NotTo(HaveOccurred())
				Expect(dirtyHeap).To(Equal(0))
			})

			It("knows when the heap has been modified", func() {
				Expect(worker.Attach()).To(Succeed())
				Expect(worker.ClearMemRefs()).To(Succeed())
				Expect(worker.Continue()).To(Succeed())

				// mallocs every 50ms
				time.Sleep(time.Millisecond * 60)
				Expect(worker.Stop()).To(Succeed())
				state, err := worker.GetState()
				Expect(err).NotTo(HaveOccurred())

				// after a bit, we expect the heap to change
				dirtyHeap, err := state.CountDirtyPages("[heap]")
				Expect(worker.Detach()).To(Succeed())

				Expect(err).NotTo(HaveOccurred())
				Expect(dirtyHeap).NotTo(Equal(0))
			})

		})

		Context("for loop stack", func() {
			BeforeEach(func() {
				targetLayer = "forloopstack"
			})

			It("can notice a variable change the stack", func() {
				Expect(worker.Attach()).To(Succeed())
				Expect(worker.ClearMemRefs()).To(Succeed())
				Expect(worker.Continue()).To(Succeed())

				// loop ticks every 50ms
				time.Sleep(time.Millisecond * 60)
				state, err := worker.GetState()
				Expect(err).NotTo(HaveOccurred())

				dirtyStack, err := state.CountDirtyPages("[stack]")
				Expect(worker.Detach()).To(Succeed())

				Expect(err).NotTo(HaveOccurred())
				Expect(dirtyStack).NotTo(Equal(0))
			})

			It("can make a copy of an area of memory", func() {
				// loop ticks every 50ms
				time.Sleep(time.Millisecond * 60)
				Expect(worker.Attach()).To(Succeed())
				defer worker.Detach()

				state, err := worker.GetState()
				Expect(err).NotTo(HaveOccurred())

				Expect(state.SavePages("[stack]")).To(Succeed())

				memSize, err := state.MemorySize("[stack]")
				Expect(err).NotTo(HaveOccurred())
				Expect(memSize).NotTo(Equal(0))
			})

			It("has no memory region changes", func() {
				Expect(worker.Attach()).To(Succeed())
				defer worker.Detach()

				state, err := worker.GetState()
				Expect(err).NotTo(HaveOccurred())
				Expect(worker.Continue()).To(Succeed())

				time.Sleep(time.Millisecond * 100)
				Expect(worker.Stop()).To(Succeed())
				changed, err := state.MemoryChanged()
				Expect(err).NotTo(HaveOccurred())
				Expect(changed).To(BeFalse())
			})

			It("has three file descriptors", func() {
				Expect(worker.Attach()).To(Succeed())
				defer worker.Detach()

				state, err := worker.GetState()
				Expect(err).NotTo(HaveOccurred())
				Expect(worker.Continue()).To(Succeed())

				Expect(len(state.GetFileDescriptors())).To(Equal(3))
			})

			It("has no changes in files", func() {
				Expect(worker.Attach()).To(Succeed())
				defer worker.Detach()

				state, err := worker.GetState()
				Expect(err).NotTo(HaveOccurred())
				Expect(worker.Continue()).To(Succeed())

				time.Sleep(time.Millisecond * 60)
				Expect(worker.Stop()).To(Succeed())
				changed, err := state.FdsChanged()
				Expect(err).NotTo(HaveOccurred())
				Expect(changed).To(BeFalse())
			})
		})

		Context("expanding heap", func() {
			BeforeEach(func() {
				targetLayer = "growingheap"
			})

			It("notices when the process changes memory regions", func() {
				Expect(worker.Attach()).To(Succeed())
				defer worker.Detach()

				state, err := worker.GetState()
				Expect(err).NotTo(HaveOccurred())
				Expect(worker.Continue()).To(Succeed())

				time.Sleep(time.Millisecond * 100)
				Expect(worker.Stop()).To(Succeed())
				changed, err := state.MemoryChanged()
				Expect(err).NotTo(HaveOccurred())
				Expect(changed).To(BeTrue())
			})
		})

		Context("opened files", func() {
			BeforeEach(func() {
				targetLayer = "fileopener"
			})

			It("notices changes in files", func() {
				Expect(worker.Attach()).To(Succeed())
				defer worker.Detach()

				state, err := worker.GetState()
				Expect(err).NotTo(HaveOccurred())
				Expect(worker.Continue()).To(Succeed())

				time.Sleep(time.Millisecond * 60)
				Expect(worker.Stop()).To(Succeed())
				changed, err := state.FdsChanged()
				Expect(err).NotTo(HaveOccurred())
				Expect(changed).To(BeTrue())
			})
		})
	})
})