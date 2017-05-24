package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	_ "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

//TODO Maybe pull these out to a config file or something?
//FOR NOW - This is where you can come to initially or iteratively update how your project is built
const (
	PROJECT_NAME  = "fights"
	BUILD_PACKAGE = "emc.com/fightsvc"
	GOLANGVERSION = "1.7"
)

var env []string
var cwd string

var (
	race    = flag.Bool("race", false, "Build race-detector version of binaries (they will run slowly)")
	verbose = flag.Bool("v", false, "Verbose mode")
	quiet   = flag.Bool("quiet", false, "Don't print anything unless there's a failure.")
	docker  = flag.Bool("docker", false, "This will do all go building in a docker container, bypassing the need to have go installed on your system (Unless the binary isn't provided).")
)

func makeInstallTools() {
	var output bytes.Buffer
	if !*quiet {
		log.Println("### Install Build Tools          ###")
	}
	var cmd *exec.Cmd

	//Verify Go Version
	verifyGoVersion()
	cmd = exec.Command("go", "install", "golang.org/x/tools/cmd/cover", "golang.org/x/tools/cmd/vet", "github.com/golang/lint/golint", "github.com/kisielk/errcheck", "github.com/t-yuki/gocover-cobertura")
	cmd.Dir = cwd
	cmd.Env = env

	if *quiet {
		cmd.Stdout = &output
		cmd.Stderr = &output
	}
	if *verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error building tools for project: %v\n%s", err, output.String())
	}

	//Copy to GOROOT + /bin/k
	cmd = exec.Command("cp", cwd+"/.tools/bin/errcheck", cwd+"/.tools/bin/gocover-cobertura", cwd+"/.tools/bin/golint", os.Getenv("GOROOT")+"/bin/")
	cmd.Dir = cwd
	cmd.Env = env
	if *quiet {
		cmd.Stdout = &output
		cmd.Stderr = &output
	}
	if *verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		log.Fatalf("Check permissions. Error moving tools for project to "+os.Getenv("GOROOT")+"/bin/: %v\n%s", err, output.String())
	}
	if !*quiet {
		log.Println("### Finished Install Build Tools ###")
	}
}

func makeGetVendorLibs() {
	var output bytes.Buffer
	if !*quiet {
		log.Println("### Get         ###")
	}
	var cmd *exec.Cmd

	//Verify Go Version
	verifyGoVersion()
	cmd = exec.Command("go", "get", "-d")
	cmd.Dir = cwd + "/src/" + BUILD_PACKAGE
	cmd.Env = env

	if *quiet {
		cmd.Stdout = &output
		cmd.Stderr = &output
	}
	if *verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error getting dependencies for project: %v\n%s", err, output.String())
	}
	if !*quiet {
		log.Println("### Finished Get ###\n")
	}
}

//Clean everything out
func makeCleanDirTarget() {
	var output bytes.Buffer
	if !*quiet {
		log.Println("### Clean          ###")
	}
	cmd := exec.Command("rm", "-rvf",
		PROJECT_NAME,
		"pkg",
		"bin",
		".vendor/pkg",
		".vendor/bin",
		".tools/pkg",
		".tools/bin",
		"pipeline",
		"profile.cov",
		"coverage.xml",
		PROJECT_NAME+".tar.gz")
	if *quiet {
		cmd.Stdout = &output
		cmd.Stderr = &output
	}
	if *verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error cleaning project: %v\n%s", err, output.String())
	}
	if !*quiet {
		log.Println("### Finished Clean ###\n")
	}
}

//Check format all relavent code
func makeCheckFormatTarget() {
	var output bytes.Buffer

	if !*quiet {
		log.Println("### Check Format          ###")
	}
	var cmd *exec.Cmd

	//Verify Go Version
	verifyGoVersion()
	goArgs := []string{"-l", cwd + "/src/."}
	cmd = exec.Command("gofmt", goArgs...)
	cmd.Env = env

	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		log.Fatalf("Error formatting project: %v\n%s", err, output.String())
	}
	if !*quiet {
		message := output.String()
		if message != "" {
			log.Fatal(message, "\nFormat is not correct.")
		}
		log.Println("### Finished Check Format ###\n")
	}
}

//Check lint all relavent code
func makeLintTarget() {
	var output bytes.Buffer

	if !*quiet {
		log.Println("### Lint          ###")
	}
	var cmd *exec.Cmd

	//Verify Go Version
	verifyGoVersion()
	goArgs := []string{cwd + "/src/..."}
	cmd = exec.Command("golint", goArgs...)
	cmd.Env = env

	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		log.Fatalf("Error formatting project: %v\n%s", err, output.String())

		cmd = exec.Command("printenv")
		cmd.Env = env

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
		if err := cmd.Run(); err != nil {
			log.Fatalf("Error formatting project: %v\n%s", err, output.String())
		}
	}
	if !*quiet {
		message := output.String()
		if message != "" {
			log.Fatal(message, "\n## Lint is not correct. ##")
		}
		log.Println("### Finished Lint ###\n")
	}
}

func makeErrTarget() {
	var output bytes.Buffer
	if !*quiet {
		log.Println("### Check Errors        ###")
	}
	var cmd *exec.Cmd

	//Verify Go Version
	verifyGoVersion()
	goArgs := []string{"emc.com/..."}
	cmd = exec.Command("errcheck", goArgs...)
	cmd.Env = env

	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		log.Fatalf("Error testing project: %v\n%s", err, output.String())
	}
	if !*quiet {
		message := output.String()
		if message != "" {
			log.Fatal(message, "\n## Errors. ##")
		}
		log.Println("### Finished Check Errors ###")
	}
}

//Format all relavent code
func makeFormatTarget() {
	var output bytes.Buffer
	if !*quiet {
		log.Println("### Format          ###")
	}
	var cmd *exec.Cmd

	//Verify Go Version
	verifyGoVersion()
	goArgs := []string{"fmt", "-x"}
	//goArgs = append(goArgs, testPackages...)
	cmd = exec.Command("go", goArgs...)
	cmd.Env = env

	if *quiet {
		cmd.Stdout = &output
		cmd.Stderr = &output
	}
	if *verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error formatting project: %v\n%s", err, output.String())
	}
	if !*quiet {
		log.Println("### Finished Format ###\n")
	}
}

//Build the artifact
//Some light reading on producing staticly linked binaries
//http://tschottdorf.github.io/linking-golang-go-statically-cgo-testing/
//https://medium.com/@kelseyhightower/optimizing-docker-images-for-static-binaries-b5696e26eb07
//http://blog.xebia.com/2014/07/04/create-the-smallest-possible-docker-container/
//In 1.4 it was broken, there is a workaround: https://github.com/golang/go/issues/9344
func makeBuildTarget() {
	var output bytes.Buffer
	if !*quiet {
		log.Println("### Build          ###")
	}
	var cmd *exec.Cmd

	args := make([]string, 0)
	args = append(args, "build", "-a", "-ldflags", "'-s'", "-installsuffix", "cgo", "-o", PROJECT_NAME, "-v")
	if *race {
		args = append(args, "-race")
	}
	args = append(args, BUILD_PACKAGE)
	//Verify Go Version
	verifyGoVersion()
	cmd = exec.Command("go", args...)
	cmd.Env = env

	if *quiet {
		cmd.Stdout = &output
		cmd.Stderr = &output
	}
	if *verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error building project: %v\n%s", err, output.String())
	}
	if !*quiet {
		log.Println("### Finished Build ###\n")
	}
}

//This guy will build the pipeline artifact
func makePipelineTarget() {
	if !*quiet {
		log.Println("### Pipeline          ###")
	}

	err := os.MkdirAll(cwd+"/pipeline/implementation/docker/"+PROJECT_NAME, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}
	if *verbose {
		log.Println("Created\t./pipeline/implementation/docker/" + PROJECT_NAME)
	}

	mdf, err := os.Create(cwd + "/pipeline/implementation/docker/" + PROJECT_NAME + "/Dockerfile")
	if err != nil {
		log.Fatalf("Failed to create Dockerfile: %v", err)
	}
	defer mdf.Close()
	mdf.Write([]byte(`FROM ubuntu:14.04
RUN apt-get update
RUN apt-get install -y net-tools
ADD ` + PROJECT_NAME + ` /` + PROJECT_NAME + `
ENV VERSION ` + getVersion() + `
ENV GIT_VERSION ` + gitVersion() + `
ENV LOG_TO_STD_ERR TRUE
ENV LOG_STD_ERR_THRESHOLD INFO
ENV LOG_VERBOSITY 3
ENTRYPOINT ["/` + PROJECT_NAME + `"]`))
	if *verbose {
		log.Println("Created\t./pipeline/implementation/docker/" + PROJECT_NAME + "/Dockerfile")
	}

	mcj, err := os.Create(cwd + "/pipeline/implementation/docker/" + PROJECT_NAME + "/container.json")
	if err != nil {
		log.Fatalf("Failed to create container.json: %v\n", err)
	}
	defer mcj.Close()

	mcj.Write([]byte(`{"external": true,"name":"` + PROJECT_NAME + `","image":"` + PROJECT_NAME + `","ports":[{"containerPort":8080,"hostPort":8080}], "livenessProbe": { "httpGet": { "path": "/heartbeat", "port": 8080 }, "initialDelaySeconds": 30, "timeoutSeconds": 2 }}`))
	if *verbose {
		log.Println("Created\t./pipeline/implementation/docker/" + PROJECT_NAME + "/container.json")
	}

	if err := exec.Command("cp", "etc>ssl>certs>ca-certificates.crt", "pipeline/implementation/docker/"+PROJECT_NAME+"/etc>ssl>certs>ca-certificates.crt").Run(); err != nil {
		log.Fatalf("Failed to copy ./etc>ssl>certs>ca-certificates.crt to ./pipeline/implementation/docker/"+PROJECT_NAME+"/etc>ssl>certs>ca-certificates.crt: %v\n", err)
	}
	if *verbose {
		log.Println("Copied\t./etc>ssl>certs>ca-certificates.crt ./pipeline/implementation/docker/" + PROJECT_NAME + "/etc>ssl>certs>ca-certificates.crt")
	}

	if err := exec.Command("cp", PROJECT_NAME, "pipeline/implementation/docker/"+PROJECT_NAME+"/"+PROJECT_NAME).Run(); err != nil {
		log.Fatalf("Failed to copy ./"+PROJECT_NAME+", ./key.pem, ./cert.pem to ./pipeline/implementation/docker/"+PROJECT_NAME+"/"+PROJECT_NAME+": %v\n", err)
	}
	if *verbose {
		log.Println("Copied\t./" + PROJECT_NAME + ", ./key.pem, ./cert.pem ./pipeline/implementation/docker/" + PROJECT_NAME + "/" + PROJECT_NAME)
	}

	if flag.Arg(0) == "deploy" {
		if err := exec.Command("cp", "coverage.xml", "pipeline/coverage.xml").Run(); err != nil {
			log.Fatalf("Failed to copy coverage.xml to ./pipeline/: %v\n", err)
		}
		if *verbose {
			log.Println("Copied\tcoverage.xml ./pipeline/coverage.xml")
		}
	}

	//		Create a file called service.json
	sjsn, err := os.Create(cwd + "/pipeline/service.json")
	if err != nil {
		log.Fatalf("Failed to create service.json: %v", err)
	}
	defer sjsn.Close()
	sjsn.Write([]byte(`{"tests":{"functional":["tests/docker/unit"]},"dependencies": []}`))
	if *verbose {
		log.Println("Created\t./pipeline/service.json")
	}

	//		Create a file called definition.raml
	drml, err := os.Create(cwd + "/pipeline/definition.raml")
	if err != nil {
		log.Fatalf("Failed to create definition.raml: %v", err)
	}
	defer drml.Close()
	drml.Write([]byte(`#`))
	if *verbose {
		log.Println("Created\t./pipeline/definition.raml")
	}
	//Pull QA from repo
	if err = exec.Command("git", "clone", "ssh://jenkins@gerrit.mozy.lab.emc.com:29418/dpc-svcs-core-iam-qa").Run(); err != nil {
		log.Println("Failed to clone QA repo. See https://confluence.dpc.lab.emc.com/x/-6OQ for solutions to this problem.", err)
	}
	//Copy Qa folder
	err = os.MkdirAll(cwd+"/pipeline/tests/docker/", os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}
	if err = exec.Command("cp", "-r", "./dpc-svcs-core-iam-qa/qa", "pipeline/tests/docker/").Run(); err != nil {
		if err = exec.Command("cp", "-r", "./qa", "pipeline/tests/docker/").Run(); err != nil {
			log.Fatalf("Failed to copy QA content to ./pipeline/tests/docker %v\n", err)
		}
	}
	if *verbose {
		log.Println("Copied\t./dpc-svcs-core-iam-qa/qa to ./pipeline/tests/docker/")
	}

	//Tar up the pipeline directory
	tarCmd := exec.Command("tar", "-pczf", "../"+PROJECT_NAME+".tar.gz", ".")
	tarCmd.Dir = cwd + "/pipeline"
	tarCmd.Run()
	if *verbose {
		log.Println("Created\t./" + PROJECT_NAME + ".tar.gz")
	}

	if !*quiet {
		log.Println("### Finished Pipeline ###\n")
	}
}

//This guy will build the docker image for the project (install on local machine)
func makeDockerImageTarget() {
	var output bytes.Buffer
	if !*quiet {
		log.Println("### Build Docker Image          ###")
	}
	version := getVersion()

	cmd := exec.Command("docker", "build", "-t", PROJECT_NAME+":"+version, ".")
	cmd.Dir = cwd + "/pipeline/implementation/docker/" + PROJECT_NAME

	if *quiet {
		cmd.Stdout = &output
		cmd.Stderr = &output
	}
	if *verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error building project docker images: %v\n%s", err, output.String())
	}
	if !*quiet {
		log.Println("### Finished Docker Image Build ###\n")
	}
}

func makeCleanDockerImageTarget() {
	var output bytes.Buffer
	if !*quiet {
		log.Println("### Clean Docker Image          ###")
	}
	version := getVersion()

	cmd := exec.Command("docker", "rmi", PROJECT_NAME+":"+version)

	if *quiet {
		cmd.Stdout = &output
		cmd.Stderr = &output
	}
	if *verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error cleaning docker image: %v\n%s", err, output.String())
	}
	if !*quiet {
		log.Println("### Finished Clean Docker Image ###\n")
	}
}

func makeRamlTarget() {
	var output bytes.Buffer
	if !*quiet {
		log.Println("### RAML          ###")
	}
	var cmd *exec.Cmd

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Couldn't get working directory")
	}

	err = os.Chdir(cwd + "/src/emc.com/dpc/iam/raml/")
	if err != nil {
		log.Fatal("Couldn't get working directory")
	}
	defer os.Chdir(cwd)

	args := make([]string, 0)
	args = append(args, "run")
	args = append(args, "main.go", "upload")
	//Verify Go Version
	verifyGoVersion()
	cmd = exec.Command("go", args...)
	cmd.Env = env

	if *quiet {
		cmd.Stdout = &output
		cmd.Stderr = &output
	}
	if *verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		log.Fatalf("Error with raml: %v\n%s", err, output.String())
	}
	if !*quiet {
		log.Println("### Finished RAML ###\n")
	}
}

func makeDockerContainerTarget() {
	//TODO Run docker run -d {projName}:{version}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	if !*quiet {
		log.Println("Make Go")
		log.Println("Your GOROOT=", os.Getenv("GOROOT"))
		log.Println("Your GOPATH=", os.Getenv("GOPATH"))
		log.Println("###\n")
	}

	var err error
	env = cleanGoEnv()
	cwd, err = os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	gopath := "GOPATH=/go:" + cwd + "/.vendor:" + cwd + "/.tools:" + cwd
	env = append(env, gopath, "CGO_ENABLED=0")

	log.Println("New ", gopath)
	log.Println("New GOROOT=", os.Getenv("GOROOT"))

	makeInstallTools()

	//We can now run specific targets
	arg := flag.Arg(0)
	if !*quiet && arg != "" {
		log.Println("Target: " + arg)
		log.Println("---\n")
	}
	switch arg {
	case "raml":
		makeRamlTarget()
	case "get":
		makeGetVendorLibs()
	case "clean":
		makeCleanDirTarget()
	case "format":
		makeCleanDirTarget()
		makeFormatTarget()
	case "test":
		makeCleanDirTarget()
		//makeTestTarget()
	case "cover":
		makeCleanDirTarget()
		//makeTestCoverTarget()
	case "lint":
		makeCleanDirTarget()
		makeLintTarget()
	case "vet":
		makeCleanDirTarget()
	case "errors":
		makeCleanDirTarget()
		makeErrTarget()
	case "commit":
		makeCleanDirTarget()
		makeCheckFormatTarget()
		//makeLintTarget()
		//makeErrTarget()
		//makeTestTarget()
		makeBuildTarget()
	case "build":
		makeCleanDirTarget()
		makeFormatTarget()
		//makeTestTarget()
		makeBuildTarget()
	case "pipeline":
		makeCleanDirTarget()
		makeFormatTarget()
		//makeTestTarget()
		makeBuildTarget()
		makePipelineTarget()
	case "deploy":
		makeCleanDirTarget()
		makeFormatTarget()
		//makeTestCoverTarget()
		makeBuildTarget()
		//testInt()
		makePipelineTarget()
		//makeRamlTarget()
	default:
		makeCleanDirTarget()
		makeFormatTarget()
		//makeTestTarget()
		makeBuildTarget()
	}
}

// getVersion returns the version of the app. Either from a VERSION file at the root,
// or from git.
func getVersion() string {
	slurp, err := ioutil.ReadFile(filepath.Join(cwd, "VERSION"))
	if err == nil {
		return strings.TrimSpace(string(slurp))
	}
	return gitVersion()
}

var gitVersionRx = regexp.MustCompile(`\b\d\d\d\d-\d\d-\d\d-[0-9a-f]{7,7}\b`)

// gitVersion returns the git version of the git repo at camRoot as a
// string of the form "yyyy-mm-dd-xxxxxxx", with an optional trailing
// '+' if there are any local uncomitted modifications to the tree.
func gitVersion() string {
	cmd := exec.Command("git", "rev-list", "--max-count=1", "--pretty=format:'%ad-%h'", "--date=short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		log.Fatalf("Error running git rev-list in %s: %v", cwd, err)
	}
	v := strings.TrimSpace(string(out))
	if m := gitVersionRx.FindStringSubmatch(v); m != nil {
		v = m[0]
	} else {
		panic("Failed to find git version in " + v)
	}
	cmd = exec.Command("git", "diff", "--exit-code")
	if err := cmd.Run(); err != nil {
		v += "+"
	}
	return v
}

func verifyGoVersion() {
	pv, _ := strconv.ParseUint(strings.Split(GOLANGVERSION, ".")[1], 10, 8)
	neededMinor := uint8(pv)
	_, err := exec.LookPath("go")
	if err != nil {
		log.Fatalf("Go doesn't appeared to be installed ('go' isn't in your PATH). Install Go 1.%c or newer.", neededMinor)
	}
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		log.Fatalf("Error checking Go version with the 'go' command: %v", err)
	}

	log.Println("Using : ", string(out))

	fields := strings.Fields(string(out))
	if len(fields) < 3 || !strings.HasPrefix(string(out), "go version ") {
		log.Fatalf("Unexpected output while checking 'go version': %q", out)
	}
	version := fields[2]
	if version == "devel" {
		return
	}
	// this check is still needed for the "go1" case.
	if len(version) < len("go1.") {
		log.Fatalf("Your version of Go (%s) is too old. make.go requires Go 1.%c or later.", version, neededMinor)
	}
	minorChar := strings.TrimPrefix(version, "go1.")[0]
	if minorChar >= neededMinor && minorChar <= '9' {
		return
	}
	log.Fatalf("Your version of Go (%s) is too old. make.go requires Go 1.%c or later.\n", version, neededMinor)
}

// cleanGoEnv returns a copy of the current environment with GOPATH and GOBIN removed.
// it also sets GOOS and GOARCH as needed when cross-compiling.
func cleanGoEnv() (clean []string) {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "GOPATH=") {
			continue
		}
		// We skip these two as well, otherwise they'd take precedence over the
		// ones appended below.
		if strings.HasPrefix(env, "GOOS=") {
			continue
		}
		if strings.HasPrefix(env, "GOARCH=") {
			continue
		}
		clean = append(clean, env)
	}
	return
}
