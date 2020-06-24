package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/rakyll/statik/fs"
	_ "github.com/yasutakatou/rabbitRPA/statik"
	//_ "./statik"
)

//FYI: https://journal.lampetty.net/entry/capturing-stdout-in-golang
type Capturer struct {
	saved         *os.File
	bufferChannel chan string
	out           *os.File
	in            *os.File
}

func main() {
	Debug := false

	args := ""

	for i, v := range os.Args {
		if strings.Index(v, "-debug") != -1 {
			Debug = true
		}
		if i > 0 {
			args += v + " "
		}
	}

	checkExeDlls()

	if Debug == true {
		fmt.Println("launch> do.exe " + args)
	}

	Execmd("do.exe " + args)

	os.Exit(0)
}

func Execmd(command string) {
	cmd := exec.Command("cmd", "/C", command)
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

//statik -src=./ -include=*.dll,*.exe
func checkExeDlls() {
	exeDlls := []string{"do.exe", "libopencv_calib3d430.dll", "libopencv_core430.dll", "libopencv_dnn430.dll", "libopencv_features2d430.dll",
		"libopencv_flann430.dll", "libopencv_highgui430.dll", "libopencv_imgcodecs430.dll", "libopencv_imgproc430.dll",
		"libopencv_objdetect430.dll", "libopencv_video430.dll", "libopencv_videoio430.dll"}

	for i := 0; i < len(exeDlls); i++ {
		if Exists(exeDlls[i]) == false {
			if makeFile(exeDlls[i]) == false {
				fmt.Println("create failure: ", exeDlls[i])
				os.Exit(1)
			}
		}
	}
}

func makeFile(filename string) bool {
	statikFS, err := fs.New()
	if err != nil {
		return false
	}

	r, err := statikFS.Open("/" + filename)
	if err != nil {
		return false
	}
	defer r.Close()

	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return false
	}

	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return false
	}

	_, err = file.Write(contents)

	if err != nil {
		return false
	}

	return true
}
