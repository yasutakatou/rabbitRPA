# rabbitRPA

### (Probably) the world's first one file RPA tool impremented by Golang!

![icon](https://github.com/yasutakatou/rabbitRPA/blob/pics/icon.png)

lovely Lepus brachyurus!<br>

# solution

You try to automation on windows gui operation, you might use RPA tool. 
after, you will setup this tool requirement for execute.
install, config file customize, and depend tool and more.
### I wonder why, RPA tool are too fat.
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
