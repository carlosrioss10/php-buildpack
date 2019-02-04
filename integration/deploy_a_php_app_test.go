package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var _ = Describe("V3 Wrapped CF PHP Buildpack", func() {
	var app *cutlass.App
	AfterEach(func() {
		DestroyApp(app)
	})

	Context("When pushing a simple PHP script app", func() {
		BeforeEach(func() {
			app = cutlass.New(filepath.Join(bpDir, "fixtures", "simple_app"))
			app.Disk = "1G"
			app.Memory = "1G"
		})

		It("uses the requested php version and runs successfully", func() {
			Expect(app.Push()).To(Succeed())
			Eventually(func() ([]string, error) { return app.InstanceStates() }, 120*time.Second).Should(Equal([]string{"RUNNING"}))
			Eventually(app.Stdout.String).Should(MatchRegexp(`.*PHP.*\d+\.\d+\.\d+.*:.*Contributing.*`))
			Eventually(app.Stdout.String).Should(ContainSubstring("OUT SUCCESS"))
		})
	})

	Context("Unbuilt buildpack (eg github)", func() {
		var bpName string

		BeforeEach(func() {
			if cutlass.Cached {
				Skip("skipping cached buildpack test")
			}

			tmpDir, err := ioutil.TempDir("", "")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(tmpDir)

			bpName = "unbuilt-v3-php"
			bpZip := filepath.Join(tmpDir, bpName+".zip")

			app = cutlass.New(filepath.Join(bpDir, "fixtures", "simple_app"))
			app.Buildpacks = []string{bpName + "_buildpack"}
			app.Disk = "1G"
			app.Memory = "1G"

			cmd := exec.Command("git", "archive", "-o", bpZip, "HEAD")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Dir = bpDir
			Expect(cmd.Run()).To(Succeed())

			Expect(cutlass.CreateOrUpdateBuildpack(bpName, bpZip, "")).To(Succeed())
		})

		AfterEach(func() {
			Expect(cutlass.DeleteBuildpack(bpName)).To(Succeed())
		})

		It("runs", func() {
			Expect(app.Push()).To(Succeed())
			Eventually(func() ([]string, error) { return app.InstanceStates() }, 120*time.Second).Should(Equal([]string{"RUNNING"}))

			Eventually(app.Stdout.String).Should(MatchRegexp(`.*PHP.*\d+\.\d+\.\d+.*:.*Contributing.*`))
			Eventually(app.Stdout.String).Should(ContainSubstring("OUT SUCCESS"))
		})
	})
})