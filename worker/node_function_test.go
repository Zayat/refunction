package worker_test

import (
	"io"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	. "github.com/ostenbom/refunction/worker"
)

var _ = Describe("Node Serverless Function Management", func() {
	var id string
	runtime := "node"
	var targetLayer string
	var worker *Worker
	var stdout *gbytes.Buffer
	var straceBuffer *gbytes.Buffer

	BeforeEach(func() {
		id = strconv.Itoa(GinkgoParallelNode())
	})

	JustBeforeEach(func() {
		var err error
		worker, err = NewWorker(id, client, runtime, targetLayer)
		Expect(err).NotTo(HaveOccurred())
		stdout = gbytes.NewBuffer()
		worker.WithStdPipes(GinkgoWriter, stdout, GinkgoWriter)

		straceBuffer = gbytes.NewBuffer()
		multiBuffer := io.MultiWriter(straceBuffer, GinkgoWriter)
		worker.WithSyscallTrace(multiBuffer)

		Expect(worker.Start()).To(Succeed())
	})

	AfterEach(func() {
		err := worker.End()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("server managed functions", func() {

		BeforeEach(func() {
			targetLayer = "serverless-function.js"
		})

		It("can load a function", func() {
			// Initiate python ready sequence
			Expect(worker.Activate()).To(Succeed())
			Expect(len(worker.GetCheckpoints())).To(Equal(1))
			Eventually(stdout).Should(gbytes.Say("started"))

			function := "function main(p) { return p }"
			Expect(worker.SendFunction(function)).To(Succeed())
			Eventually(stdout).Should(gbytes.Say("{\"type\":\"function_loaded\",\"data\":true}"))
		})

		It("can get a request response", func() {
			Expect(worker.Activate()).To(Succeed())

			function := "function main(p) { return p }"
			Expect(worker.SendFunction(function)).To(Succeed())

			request := "jsonstring"
			response, err := worker.SendRequest(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(Equal(request))
		})

		It("can get an object request response", func() {
			Expect(worker.Activate()).To(Succeed())

			function := "function main(p) { return p }"
			Expect(worker.SendFunction(function)).To(Succeed())

			request := map[string]interface{}{
				"greatkey": "nicevalue",
			}
			response, err := worker.SendRequest(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(Equal(request))
		})

		It("can get several request responses", func() {
			Expect(worker.Activate()).To(Succeed())

			function := "function main(p) { return p }"
			Expect(worker.SendFunction(function)).To(Succeed())
			Eventually(stdout).Should(gbytes.Say("{\"type\":\"function_loaded\",\"data\":true}"))

			request := "jsonstring"
			response, err := worker.SendRequest(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(Equal(request))

			request = "anotherstring"
			response, err = worker.SendRequest(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(Equal(request))

			request = "whateverstring"
			response, err = worker.SendRequest(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(Equal(request))
		})

		It("can restore and change function", func() {
			Expect(worker.Activate()).To(Succeed())

			function := "function main(p) { return p }"
			Expect(worker.SendFunction(function)).To(Succeed())

			request := "jsonstring"
			response, err := worker.SendRequest(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(Equal(request))

			Expect(worker.Restore()).To(Succeed())

			function = "function main(p) { return 'unrelated' }"
			Expect(worker.SendFunction(function)).To(Succeed())

			request = "anotherstring"
			response, err = worker.SendRequest(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(Equal("unrelated"))
		})

		It("is resiliant to improper function loads", func() {
			Expect(worker.Activate()).To(Succeed())

			// python for example
			function := "def main(req):\n  print(req)\n  return req"
			err := worker.SendFunction(function)
			Expect(err).NotTo(BeNil())
			Eventually(stdout).Should(gbytes.Say("{\"type\":\"function_loaded\",\"data\":false}"))

			function = "function main(params) {\n    return params || {};\n}\n"
			Expect(worker.SendFunction(function)).To(Succeed())
			Eventually(stdout).Should(gbytes.Say("{\"type\":\"function_loaded\",\"data\":true}"))

			request := "jsonstring"
			response, err := worker.SendRequest(request)
			Expect(err).NotTo(HaveOccurred())
			Expect(response).To(Equal(request))
		})
	})
})
