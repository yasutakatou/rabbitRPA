# rabbitRPA

### (Probably) the world's first one file RPA tool impremented by Golang!

![icon](https://github.com/yasutakatou/rabbitRPA/blob/pics/icon.png)

lovely Lepus brachyurus!<br>

### So far, it only works on Windows.

# demo

![demo](https://github.com/yasutakatou/rabbitRPA/blob/pics/rabbitRPA.gif)

# solution

You try to automation on windows gui operation, you might use RPA tool. 
after, you will setup this tool requirement for execute.
install, config file customize, and depend tool and more.
### I wonder why, RPA tools are too fat.
I want do old installer to automate, only once.
But, RPA set up is a pain in the ass. I think needs, light weight RPA tool.

# features

 - **one binary (it's not perfect)**
 - light weight
 - very simple use
 - (Of course) free!

# structure

1. one execute file include all binary and dll files.
2. when execute this, extract execute file for RPA, and OpenCV dll files.
3. and then, execute RPA binary depend for OpenCV, Your operation recorded or replayed!

# algorithm

this tool have two functions roughly. It's Record or Replay.<br>

## Record

When no option, tool is Record mode.<br>

```
rabbitRPA.exe
```

tool executed and create require files, after follow message.


```
 - - recording start! you want to end this mode, key press ascii code (%d) - -
```

until input exit key code, your operation recorded.<br>
(default exit key code is **27[Escape Key]**.)<br>

After input exit key code, your operation is recorded tsv file and captures.<br>

tsv file include target window title, move value mouse, click position.<br>
captures include all screen, and target window capture.<br>
*tool calculates the capture's difference.*<br>
*Therefore, If target windows moved another position, that position adjust.*<br>

<br>

## Replay

Replay the operation using the file you just used. set you options. <br>

```
rabbitRPA.exe -replay -import=output.tsv
```

*"-replay"* is replay mode option.<br>
*"-import"* is exported tsv file for record mode.<br>

### your record will replay!

[See here for other options.](https://github.com/yasutakatou/rabbitRPA#options)

# installation

download binary from [release page](https://github.com/yasutakatou/rabbitRPA/releases).<br>
save binary file, copy to entryed execute path directory.<br>

## another

this tool depend gocv and statik.<br>

[GoCV](https://github.com/hybridgroup/gocv)<br>
[statik](https://github.com/rakyll/statik)

step1. you install gocv(and opencv)

See below.

https://github.com/hybridgroup/gocv#windows

```
win_build_opencv.cmd
```

step2. "statik" set up

```
go get github.com/rakyll/statik
```

step3. clone this repository

```
git clone https://github.com/yasutakatou/rabbitRPA
cd rabbitRPA
```

step4. copy opencv dlls and converted by statik

Copy the dll from the opencv folder. (total 11 files)<br>
ex) **C:\opencv\build\install\x64\mingw\bin** to **.\rabbitRPA**<br>

```
libopencv_calib3d430.dll
libopencv_core430.dll
libopencv_dnn430.dll
libopencv_features2d430.dll
libopencv_flann430.dll
libopencv_highgui430.dll
libopencv_imgcodecs430.dll
libopencv_imgproc430.dll
libopencv_objdetect430.dll
libopencv_video430.dll
libopencv_videoio430.dll
```

Build the RPA tool.

```
go build do.go
```

converting by use statik.

```
cd rabbitRPA
statik -src=./ -include=*.dll,*.exe
```

Finally, package everything up.

```
go build rabbitRPA.go
```

## uninstall

```
delete that binarys.
```

del or rm command. *(it's simple!)*

# options

*note) this options give to RPA binary(do.exe) as is. therefore, do.exe's options same too.*

|option name|default value|detail|
|:---|:---|:---|
-list|false|listing window titiles and exit<br>use to search target window title.|
-replay|false|replay mode (true is enable)|
-record|true|recording mode (true is enable)<br>default is on(recording mode)|
-export|output.tsv|export file name<br>**If exists same file name, it's overwriten**.|
-import|input.tsv|import file name|
-debug|false|debug mode (true is enable)|
-threshold|0.5|same window threshold<br>**The lower the value, the more, select large difference**.<br>If screen size is more larger when it's replay, try value lower.|
-move|50|mouse move record threshold<br>record the mouse move par this value.<br>When value is lower, recording often.|
-try|10|error and try counter<br>**In case of wait next screen a while, set value larger**.|
-wait|100|loop wait Millisecond|
-exitCode|27|recording mode to exit ascii key code<br>ascii code, [please refer this site](http://www9.plala.or.jp/sgwr-t/c_sub/ascii.html).|
-tmpDir|tmp|temporary directory name<br>save captures to this directory.<br>**If exists same directory name, it's overwriten**.|

# Problem

 - If exists same title window, doing on smaller window handle.

Because search partial of window title.<br>
When focus some tool, this tool's window title auto changed, this tool can't focus that.<br>
Then, search partial.<br>

# copyright

(Probably) the world's first one file RPA tool impremented by Golang!<br>
Copyright (c) 2020 yasutakatou

# license

3-clause BSD License<br>
 and<br>
Apache License Version 2.0<br>
