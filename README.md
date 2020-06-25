# rabbitRPA

### (Probably) the world's first one file RPA tool impremented by Golang!

![icon](https://github.com/yasutakatou/rabbitRPA/blob/pics/icon.png)

lovely Lepus brachyurus!<br>

# solution

You try to automation on windows gui operation, you might use RPA tool. 
after, you will setup this tool requirement for execute.
install, config file customize, and depend tool and more.
### I wonder why, RPA tools are too fat.
I want do old installer to automate, only once.
But, RPA set up is a pain in the ass. I think needs, light weight RPA tool.

# features

 - one binary (it's not perfect)
 - light weight
 - very simple use
 - (Of course) free!

# structure

1. one execute file include all binary and dll files.
2. when execute this, extract execute file for RPA, and OpenCV dll files.
3. and then, execute RPA binary depend for OpenCV, Your operation recorded or replayed!

# installation

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

# options

*note) this options give to RPA binary(do.exe) as is. therefore, do.exe's options same too.*

|option name|default value|detail|
|:---|:---|:---|
-list|false|listing window titiles and exit|
-replay|false|replay mode (true is enable)|
-record|true|recording mode (true is enable)|
-export|output.tsv|export file name|
-import|input.tsv|import file name|
-debug|false|debug mode (true is enable)|
-threshold|0.5|same window threshold|
-move|50|mouse move record threshold|
-try|10|error and try counter|
-wait|100|loop wait Millisecond|
-exitCode|27|recording mode to exit ascii key code|
-tmpDir|tmp|temporary directory name|

# copyright

(Probably) the world's first one file RPA tool impremented by Golang!<br>
Copyright (c) 2020 yasutakatou

# license

3-clause BSD License<br>
 and<br>
Apache License Version 2.0<br>
