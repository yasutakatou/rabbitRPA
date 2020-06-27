/**
 * (Probably) the world's first one file RPA tool impremented by Golang!
 *
 * @author    yasutakatou
 * @copyright 2020 yasutakatou
 * @license   https://www.apache.org/licenses/LICENSE-2.0 Apache-2.0 , 3-clause BSD License
 */
package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/go-vgo/robotgo"
	"github.com/kbinani/screenshot"
	hook "github.com/robotn/gohook"
	"github.com/yasutakatou/string2keyboard"
	"gocv.io/x/gocv"
	"golang.org/x/image/bmp"
)

type (
	HANDLE uintptr
	HWND   HANDLE
)

type RECTdata struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

var (
	user32                  = syscall.MustLoadDLL("user32.dll")
	procEnumWindows         = user32.MustFindProc("EnumWindows")
	procGetWindowTextW      = user32.MustFindProc("GetWindowTextW")
	procSetActiveWindow     = user32.MustFindProc("SetActiveWindow")
	procSetForegroundWindow = user32.MustFindProc("SetForegroundWindow")
	procGetForegroundWindow = user32.MustFindProc("GetForegroundWindow")
	procGetWindowRect       = user32.MustFindProc("GetWindowRect")

	rs1Letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	History = []historyData{}

	Hashs = []hashData{}

	Debug             bool
	LiveExitAsciiCode int
	preWindow         string
	aftWindow         string
	LiveRawcodeChar   string
	sameThreshold     float32
	tryCounter        int
	waitSeconds       int
	TEMPDir           string
	prevDir           string
	tempX             int
	tempY             int
	moveThreshold     int
)

type historyData struct {
	Device string `json:"Device"`
	Pre    string `json:"Pre"`
	Aft    string `json:"Aft"`
	Params string `json:"Params"`
}

type hashData struct {
	Filename string `json:"Filename"`
	Hash     string `json:"Hash"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
	LiveRawcodeChar = "~"
	preWindow = ""
	aftWindow = ""
}

func main() {
	_List := flag.Bool("list", false, "[-list=listing window titiles and exit]")
	_Replay := flag.Bool("replay", false, "[-replay=replay mode (true is enable)]")
	_Record := flag.Bool("record", true, "[-record=recording mode (true is enable)]")
	_exportFile := flag.String("export", "output.tsv", "[-export=export file name]")
	_importFile := flag.String("import", "input.tsv", "[-import=import file name]")
	_DEBUG := flag.Bool("debug", false, "[-debug=debug mode (true is enable)]")
	_threshold := flag.Float64("threshold", 0.1, "[-threshold=same window threshold]")
	_move := flag.Float64("move", 50, "[-move=mouse move record threshold]")
	_try := flag.Int("try", 10, "[-try=error and try counter]")
	_wait := flag.Int("wait", 250, "[-wait=loop wait Millisecond]")
	_exitCode := flag.Int("exitCode", 27, "[-exitCode=recording mode to exit ascii key code]")
	_tmpDir := flag.String("tmpDir", "tmp", "[-tmpDir=temporary directory name]")

	flag.Parse()

	if *_List == true {
		getHwndToTitle(0, true)
		os.Exit(0)
	}

	Debug = bool(*_DEBUG)
	sameThreshold = float32(*_threshold)
	tryCounter = int(*_try)
	waitSeconds = int(*_wait)
	LiveExitAsciiCode = int(*_exitCode)
	moveThreshold = int(*_move)

	prevDir, _ = filepath.Abs(".")

	TEMPDir = prevDir + "\\" + string(*_tmpDir) + "\\"
	if Exists(TEMPDir) == true {
		if *_Record == true && *_Replay == false {
			if err := os.RemoveAll(TEMPDir); err != nil {
				fmt.Println(err)
			}
		}
	}

	if err := os.MkdirAll(TEMPDir, 0777); err != nil {
		fmt.Println(err)
	}

	os.Chdir(TEMPDir)

	if *_DEBUG == true {
		fmt.Println(" - - - options - - - ")
		fmt.Println("list: ", *_List)
		fmt.Println("replay: ", *_Replay)
		fmt.Println("record: ", *_Record)
		fmt.Println("export: ", *_exportFile)
		fmt.Println("import: ", *_importFile)
		fmt.Println("debug: ", Debug)
		fmt.Println("threshold: ", sameThreshold)
		fmt.Println("move: ", moveThreshold)
		fmt.Println("try: ", tryCounter)
		fmt.Println("wait: ", waitSeconds)
		fmt.Println("exitCode: ", LiveExitAsciiCode)
		fmt.Println("tmpDir: ", TEMPDir)
		fmt.Println(" - - - - - - - - - ")
	}

	switch *_Record {
	case true:
		if *_Replay == true {
			replayMode(prevDir + "\\" + *_importFile)
		} else {
			recordingMode(*_exportFile)
		}
	case false:
		if *_Replay == false {
			recordingMode(*_exportFile)
		} else {
			replayMode(prevDir + "\\" + *_importFile)
		}
	}

	os.Chdir(prevDir)
	os.Exit(0)
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func ImportHistory(params string) bool {
	if len(params) == 0 {
		return false
	}

	file, err := os.Open(params)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer file.Close()

	History = nil
	s := bufio.NewScanner(file)
	for s.Scan() {
		strs := strings.Split(s.Text(), "\t")
		if len(strs) != 4 {
			fmt.Println("Error: your tsv file broken")
			History = nil
			return false
		}
		History = append(History, historyData{Device: strs[0], Pre: strs[1], Aft: strs[2], Params: strs[3]})
	}
	fmt.Println("importFile: ", params)
	return true
}

func replayMode(importFile string) {
	if Exists(importFile) == false {
		fmt.Println("not found: ", importFile)
	}
	ImportHistory(importFile)

	for i := 0; i < len(History); i++ {
		fmt.Println("Device: ", History[i].Device, "Pre: ", History[i].Pre, "Aft: ", History[i].Aft, "Params: ", History[i].Params)
		if setTargetWindow(i) == false {
			return
		}
		switch History[i].Device {
		case "key":
			SendKey(History[i].Params)
		case "click", "move":
			strs := strings.Split(History[i].Params, ";")

			origFilename := getNowFilename()
			if Exists(origFilename) == false {
				break
			}
			matImage := gocv.IMRead(origFilename, gocv.IMReadGrayScale)

			targetX := 0
			targetY := 0

			partImage := ""
			if History[i].Device == "click" {
				partImage = strs[1]
				targetX, _ = strconv.Atoi(strs[2])
				targetY, _ = strconv.Atoi(strs[3])
			} else {
				partImage = strs[0]
				targetX, _ = strconv.Atoi(strs[1])
				targetY, _ = strconv.Atoi(strs[2])
			}

			if Debug == true {
				fmt.Println("origFilename: ", origFilename, " partImage: ", partImage)
			}

			if Exists(partImage) == false {
				break
			}
			matTemplate := gocv.IMRead(partImage, gocv.IMReadGrayScale)

			matResult := gocv.NewMat()
			mask := gocv.NewMat()
			gocv.MatchTemplate(matImage, matTemplate, &matResult, gocv.TmCcoeffNormed, mask)
			mask.Close()
			minConfidence, maxConfidence, minLoc, maxLoc := gocv.MinMaxLoc(matResult)

			if Debug == true {
				fmt.Println("mouseLocate: ", minConfidence, maxConfidence, minLoc, maxLoc)
			}

			if err := os.Remove(origFilename); err != nil {
				fmt.Println(err)
			}

			if maxConfidence > sameThreshold {
				actionX := maxLoc.X + targetX
				actionY := maxLoc.Y + targetY
				robotgo.MoveMouseSmooth(actionX, actionY, 0.01, 0.01)

				if History[i].Device == "click" {
					if strs[4] == "1" {
						robotgo.MouseClick("left", false)
					} else {
						robotgo.MouseClick("right", false)
					}
				}
			} else {
				if Debug == true {
					fmt.Println("under same Threshold: ", sameThreshold, " > ", maxConfidence)
				}
			}
		}
		time.Sleep(time.Duration(waitSeconds) * time.Millisecond)
	}
}

func setTargetWindow(i int) bool {
	if len(History[i].Pre) > 0 {
		if targetHwnd := FocusWindow(History[i].Pre, Debug); ChangeTarget(targetHwnd) == false {
			fmt.Println("not found title: ", History[i].Pre)
			if len(History[i].Aft) > 0 {
				if targetHwnd := FocusWindow(History[i].Aft, Debug); ChangeTarget(targetHwnd) == false {
					fmt.Println("not found title: ", History[i].Aft)
					return false
				}
			}
		}
	}
	return true
}

func getNowFilename() string {
	origFilename := ""
	nowFilename := RandStr(8) + ".bmp"

	n := screenshot.NumActiveDisplays()
	if n <= 0 {
		panic("Active display not found")
	}

	var all image.Rectangle = image.Rect(0, 0, 0, 0)

	for i := 0; i < n; i++ {
		bounds := screenshot.GetDisplayBounds(i)
		all = bounds.Union(all)
	}

	bitmap := robotgo.CaptureScreen(all.Min.X, all.Min.Y, all.Dx(), all.Dy())
	robotgo.SaveBitmap(bitmap, nowFilename)
	tmpHash := calcHash(nowFilename)

	hFlag := -1

	for i := 0; i < len(Hashs); i++ {
		if Hashs[i].Hash == tmpHash {
			hFlag = i
			break
		}
	}

	if hFlag == -1 {
		Hashs = append(Hashs, hashData{Filename: nowFilename, Hash: tmpHash})
		origFilename = nowFilename
	} else {
		origFilename = Hashs[hFlag].Filename
		if err := os.Remove(origFilename); err != nil {
			fmt.Println(err)
		}
	}
	return origFilename
}

func SendKey(doCmd string) bool {
	if Debug == true {
		fmt.Printf("KeyInput: ")
	}

	cCtrl := false
	cAlt := false

	if strings.Index(doCmd, "ctrl+") != -1 {
		cCtrl = true
		doCmd = strings.Replace(doCmd, "ctrl+", "", 1)
		if Debug == true {
			fmt.Printf("ctrl+")
		}
	}

	if strings.Index(doCmd, "alt+") != -1 {
		cAlt = true
		doCmd = strings.Replace(doCmd, "alt+", "", 1)
		if Debug == true {
			fmt.Printf("alt+")
		}
	}

	string2keyboard.KeyboardWrite(doCmd, cCtrl, cAlt, LiveRawcodeChar)
	if Debug == true {
		fmt.Println(doCmd, cCtrl, cAlt)
	}
	return true
}

func saveToBmp(img *image.RGBA, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	bmp.Encode(file, img)
}

func searchHash(targetHash string, allPartFlag bool) (string, string) {
	//all: 1, part 3, move 1
	for i := 0; i < len(History); i++ {
		if History[i].Device == "click" || History[i].Device == "move" {
			strs := strings.Split(History[i].Params, ";")
			if allPartFlag == true && targetHash == strs[1] {
				return strs[0], strs[1]
			}
			if allPartFlag == false && targetHash == strs[3] {
				return strs[2], strs[3]
			}
		}
	}
	return "", ""
}

func CaptureCase(filename, target string, setHwnd uintptr, hFlag bool) (string, string) {
	var rect RECTdata

	if target == "" && setHwnd == 0 {
		n := screenshot.NumActiveDisplays()
		if n <= 0 {
			panic("Active display not found")
		}

		var all image.Rectangle = image.Rect(0, 0, 0, 0)

		for i := 0; i < n; i++ {
			bounds := screenshot.GetDisplayBounds(i)
			all = bounds.Union(all)
		}

		bitmap := robotgo.CaptureScreen(all.Min.X, all.Min.Y, all.Dx(), all.Dy())
		robotgo.SaveBitmap(bitmap, filename)
	} else {
		ChangeTarget(setHwnd)
		GetWindowRect(HWND(setHwnd), &rect, Debug)
		bitmap := robotgo.CaptureScreen(int(rect.Left), int(rect.Top), int(rect.Right-rect.Left), int(rect.Bottom-rect.Top))
		robotgo.SaveBitmap(bitmap, filename)
	}

	tmpHash := calcHash(filename)

	if _, Hash := searchHash(tmpHash, hFlag); Hash == "" {
		return filename, tmpHash
	}

	if err := os.Remove(filename); err != nil {
		fmt.Println(err)
	}
	return "", tmpHash
}

func calcHash(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func getHwndToTitle(targetHwnd uintptr, listMode bool) string {
	lists := ListWindow(Debug)
	for i := 0; i < len(lists); i++ {
		if listMode == true {
			fmt.Println(lists[i])
		} else {
			strs := strings.Split(lists[i], " : ")
			if strings.Index(fmt.Sprintf("%x", targetHwnd), strs[1]) != -1 {
				return strs[0]
			}
		}
	}
	return ""
}

func ChangeTarget(setHwnd uintptr) bool {
	breakCounter := tryCounter

	for {
		if Debug == true {
			fmt.Printf("tryCounter: %d\n", breakCounter)
		}

		if setHwnd != GetWindow("GetForegroundWindow", Debug) {
			SetActiveWindow(HWND(setHwnd), Debug)
		} else {
			return true
		}
		breakCounter = breakCounter - 1
		if breakCounter < 0 {
			return false
		}
		time.Sleep(time.Duration(waitSeconds) * time.Millisecond)
	}
}

func recordingMode(exportFile string) {
	fmt.Printf(" - - recording start! you want to end this mode, key press ascii code (%d) - - \n", LiveExitAsciiCode)

	altFlag := 0
	actFlag := false
	var bufStrs uint16
	bufStrs = 0

	EvChan := hook.Start()
	defer hook.End()

	for ev := range EvChan {
		strs := ""

		if actFlag == true {
			if ev.Kind == 3 { //KeyDown = 3
				bufStrs, strs = keyDown(altFlag, int(ev.Rawcode), strs, string(ev.Keychar), bufStrs)
				preWindow = aftWindow
				aftWindow = getHwndToTitle(GetWindow("GetForegroundWindow", false), false)
			}

			if ev.Kind == 4 || ev.Kind == 5 { //KeyHold = 4,KeyUp = 5
				altFlag = keyHoldUp(int(ev.Rawcode), int(ev.Kind), bufStrs, exportFile)
				if altFlag == 256 {
					return
				}
				preWindow = aftWindow
				aftWindow = getHwndToTitle(GetWindow("GetForegroundWindow", false), false)
			}

			if ev.Kind == 9 { //MouseMove  = 9
				if moveValCheck(int(ev.X), int(ev.Y)) == true {
					addMouseMove(int(ev.X), int(ev.Y))
				}
				preWindow = aftWindow
				aftWindow = getHwndToTitle(GetWindow("GetForegroundWindow", false), false)
			}
		}

		if ev.Kind == 7 { //MouseHold
			if actFlag == false {
				actFlag = true
				aftWindow = getHwndToTitle(GetWindow("GetForegroundWindow", false), false)
			}
			addMouseAction(int(ev.Button), int(ev.X), int(ev.Y))
			preWindow = aftWindow
			aftWindow = getHwndToTitle(GetWindow("GetForegroundWindow", false), false)
		}

	}
}

func moveValCheck(evX, evY int) bool {
	if tempX-int(evX) > moveThreshold {
		return true
	}
	if tempY-int(evY) > (moveThreshold / 2) {
		return true
	}
	if int(evX)-tempX > moveThreshold {
		return true
	}
	if int(evY)-tempY > (moveThreshold / 2) {
		return true
	}
	return false
}

func keyDown(altFlag, Rawcode int, strs, keyChar string, bufStrs uint16) (uint16, string) {
	if altFlag == 0 {
		switch Rawcode {
		case 8:
			strs = "\\b"
		case 9:
			strs = "\\t"
		case 13:
			strs = "\\n"
		default:
			strs = keyChar
		}
	} else {
		switch altFlag {
		case 162:
			strs = "ctrl+" + keyChar
		case 164:
			strs = "alt+" + keyChar
		}
	}
	bufStrs = uint16(Rawcode)
	addHistory("key", strs)
	return bufStrs, strs
}

func keyHoldUp(Rawcode, Kind int, bufStrs uint16, exportFile string) int {
	altFlag := 0

	switch Rawcode {
	case 162, 164: //Ctrl,Alt
		if Kind == 4 {
			altFlag = Rawcode
		} else {
			altFlag = 0
		}
	case LiveExitAsciiCode: //Default Escape
		ExportHistory(exportFile)
		return 256
	case 160:
	default:
		if Kind == 5 && int(bufStrs) != Rawcode {
			addHistory("key", "\\"+LiveRawcodeChar+strconv.Itoa(Rawcode)+LiveRawcodeChar)
		}
	}
	return altFlag
}

func addMouseMove(evX, evY int) {
	var rect RECTdata

	moveFileName := RandStr(8) + ".bmp"

	currentHwnd := GetWindow("GetForegroundWindow", false)
	GetWindowRect(HWND(currentHwnd), &rect, false)
	moveFileName, moveHash := CaptureCase(moveFileName, "", currentHwnd, true)

	if moveFileName == "" {
		moveFileName, _ = searchHash(moveHash, true)
	}

	if evX > 0 && evX > int(rect.Left) && evX < int(rect.Right) {
		if evY > 0 && evX > int(rect.Top) && evY < int(rect.Bottom) {

			targetX := int(evX) - int(rect.Left)
			targetY := int(evY) - int(rect.Top)

			//targetImg,targetImgHash,x,y
			addHistory("move", moveFileName+";"+moveHash+";"+strconv.Itoa(targetX)+";"+strconv.Itoa(targetY))
			tempX = evX
			tempY = evY
		}
	}
}

func addMouseAction(evButton, evX, evY int) {
	targetX := 0
	targetY := 0

	allFilename := RandStr(8) + ".bmp"
	allFilename, allHash := CaptureCase(allFilename, "", 0, true)

	if allFilename == "" {
		allFilename, _ = searchHash(allHash, true)
	}

	partFileName := RandStr(8) + ".bmp"
	currentHwnd := GetWindow("GetForegroundWindow", Debug)
	partFileName, partHash := CaptureCase(partFileName, "", currentHwnd, false)

	if partFileName == "" {
		partFileName, _ = searchHash(partHash, false)
	}

	if Exists(allFilename) == false {
		return
	}
	matImage := gocv.IMRead(allFilename, gocv.IMReadGrayScale)

	if Exists(partFileName) == false {
		return
	}
	matTemplate := gocv.IMRead(partFileName, gocv.IMReadGrayScale)
	matResult := gocv.NewMat()
	mask := gocv.NewMat()

	gocv.MatchTemplate(matImage, matTemplate, &matResult, gocv.TmCcoeffNormed, mask)
	mask.Close()
	minConfidence, maxConfidence, minLoc, maxLoc := gocv.MinMaxLoc(matResult)

	if Debug == true {
		fmt.Println(minConfidence, maxConfidence, minLoc, maxLoc)
	}

	targetX = int(evX) - maxLoc.X
	targetY = int(evY) - maxLoc.Y

	tempX = evX
	tempY = evY

	//allImg,allImageHash,targetImg,targetImgHash,x,y,clickType
	addHistory("click", allFilename+";"+allHash+";"+partFileName+";"+partHash+";"+strconv.Itoa(targetX)+";"+strconv.Itoa(targetY)+";"+strconv.Itoa(evButton))
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func addHistory(device, strs string) {
	if len(strs) > 0 {
		History = append(History, historyData{Device: device, Pre: preWindow, Aft: aftWindow, Params: strs})
		if Debug == true {
			fmt.Println("liveRecord: ", strs)
		}
	}
}

func ListWindow(Debug bool) []string {
	var rect RECTdata

	ret := []string{}

	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		b := make([]uint16, 200)
		_, err := GetWindowText(h, &b[0], int32(len(b)))
		if err != nil {
			return 1
		}

		GetWindowRect(HWND(h), &rect, Debug)
		if rect.Left != 0 || rect.Top != 0 || rect.Right != 0 || rect.Bottom != 0 {
			if Debug == true {
				fmt.Printf("Window Title '%s' window: handle=0x%x\n", syscall.UTF16ToString(b), h)
				if rect.Left != 0 || rect.Top != 0 || rect.Right != 0 || rect.Bottom != 0 {
					fmt.Printf("window rect: ")
					fmt.Println(rect)
				}
			}
			ret = append(ret, fmt.Sprintf("%s : %x", syscall.UTF16ToString(b), h))
		}
		return 1
	})
	EnumWindows(cb, 0)
	return ret
}

func matchCheck(stra, strb string) bool {
	if strings.Index(strb, stra) != -1 {
		return true
	} else {
		if strings.Index(stra, strb) != -1 {
			return true
		}
	}
	return false
}

func FocusWindow(title string, Debug bool) uintptr {
	var hwnd syscall.Handle
	var rect RECTdata

	cb := syscall.NewCallback(func(h syscall.Handle, p uintptr) uintptr {
		b := make([]uint16, 200)
		_, err := GetWindowText(h, &b[0], int32(len(b)))
		if err != nil {
			return 1
		}

		if Debug == true {
			fmt.Printf("EnumWindows Search '%s' window: handle=0x%x\n", syscall.UTF16ToString(b), h)
		}

		if matchCheck(title, syscall.UTF16ToString(b)) == true {
			if Debug == true {
				fmt.Printf("Found! window: '%s' handle=0x%x\n", syscall.UTF16ToString(b), h)
			}
			GetWindowRect(HWND(h), &rect, Debug)
			fmt.Println()
			if int(rect.Right-rect.Left) > 0 && int(rect.Bottom-rect.Top) > 0 {
				hwnd = h
				return 0
			}
		}
		return 1
	})
	EnumWindows(cb, 0)
	return uintptr(hwnd)
}

func GetWindow(funcName string, Debug bool) uintptr {
	hwnd, _, _ := syscall.Syscall(procGetForegroundWindow.Addr(), 6, 0, 0, 0)
	if Debug == true {
		fmt.Printf("currentWindow: handle=0x%x.\n", hwnd)
	}
	return hwnd
}

func SetActiveWindow(hwnd HWND, Debug bool) {
	if Debug == true {
		fmt.Printf("SetActiveWindow: handle=0x%x.\n", hwnd)
	}
	syscall.Syscall(procSetActiveWindow.Addr(), 4, uintptr(hwnd), 0, 0)
	syscall.Syscall(procSetForegroundWindow.Addr(), 5, uintptr(hwnd), 0, 0)
}

func GetWindowRect(hwnd HWND, rect *RECTdata, Debug bool) (err error) {
	r1, _, e1 := syscall.Syscall(procGetWindowRect.Addr(), 7, uintptr(hwnd), uintptr(unsafe.Pointer(rect)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func GetWindowText(hwnd syscall.Handle, str *uint16, maxCount int32) (len int32, err error) {
	r0, _, e1 := syscall.Syscall(procGetWindowTextW.Addr(), 3, uintptr(hwnd), uintptr(unsafe.Pointer(str)), uintptr(maxCount))
	if len = int32(r0); len == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func EnumWindows(enumFunc uintptr, lparam uintptr) (err error) {
	r1, _, e1 := syscall.Syscall(procEnumWindows.Addr(), 2, uintptr(enumFunc), uintptr(lparam), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func RandStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = rs1Letters[rand.Intn(len(rs1Letters))]
	}
	return string(b)
}

func ExportHistory(filename string) bool {
	file, err := os.Create(prevDir + "\\" + filename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer file.Close()

	for i := 0; i < len(History); i++ {
		strs := ""
		if History[i].Device == "click" {
			stra := strings.Split(History[i].Params, ";")
			strs = History[i].Device + "\t" + History[i].Pre + "\t" + History[i].Aft + "\t" + stra[0] + ";" + stra[2] + ";" + stra[4] + ";" + stra[5] + ";" + stra[6]
		} else if History[i].Device == "move" {
			stra := strings.Split(History[i].Params, ";")
			strs = History[i].Device + "\t" + History[i].Pre + "\t" + History[i].Aft + "\t" + stra[0] + ";" + stra[2] + ";" + stra[3]
		} else {
			strs = History[i].Device + "\t" + History[i].Pre + "\t" + History[i].Aft + "\t" + History[i].Params
		}

		if Debug == true {
			fmt.Printf("[%3d]: %s\n", i+1, strs)
		}

		if _, err = file.WriteString(strs + "\n"); err != nil {
			fmt.Println(err)
			return false
		}
	}

	fmt.Println("exportFile: ", filename)
	return true
}
