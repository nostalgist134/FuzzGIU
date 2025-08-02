虽然不知道会不会有人看，但是我还是写一个readme吧

# 项目介绍

FuzzGIU 是一款基于 Go 语言开发的webfuzz工具，灵感来源于ffuf与burp intruder。支持多种协议（HTTP/HTTPS、WS/WSS、DNS 等），适用于 Web 漏洞扫描、API 测试等场景。

## 安装

安装fuzzGIU可以选择从源码编译

``````shell
git clone github.com/nostalgist134/FuzzGIU.git
cd FuzzGIU
go build
``````

或者到release中下载编译好的可执行文件

# 使用须知

本项目中所涉及的技术、思路和工具仅供以学习交流使用，任何人不得将其用于非法用途以及盈利等目的，否则后果自行承担。

# 使用方法

fuzzGIU完整的帮助信息如下

```powershell
PS H:\tools\fuzz\FuzzGIU> .\FuzzGIU.exe -h
Usage of H:\tools\fuzz\FuzzGIU\FuzzGIU.exe:
        H:\tools\fuzz\FuzzGIU\FuzzGIU.exe [options]
options are shown below. when fuzzGIU is executed without any args,
it will init and create plugin directory

[general settings]
  -d    request data
  -delay        delay between each job submission(millisecond) (default: 0)
  -r    request file
  -t    routine pool size (default: 64)
  -timeout      timeout(second) (default: 10)
  -u    url to giu

[matcher settings]
  -mc   match status code from response (default: 200,204,301,302,307,401,403,405,500)
  -ml   match amount of lines in response
  -mmode        matcher set operator (default: or)
  -mr   match regexp
  -ms   match response size
  -mt   match time(millisecond) to the first response byte
  -mw   match amount of words in response

[filter settings]
  -fc   filter status code from response
  -fl   filter amount of lines in response
  -fmode        filter set operator (default: or)
  -fr   filter regexp
  -fs   filter response size
  -ft   filter time(millisecond) to the first response byte
  -fw   filter amount of words in response

[HTTP settings]
  -F    follow redirects (default: false)
  -H    http headers to be used
  -X    http method (default: GET)
  -b    Cookies
  -http2        force http2 (default: false)
  -s    force https (default: false)
  -x    proxies

[payload settings]
  -mode mode for keywords used, basically the same as those in burp suite (default: clusterbomb)
  -pl-gen       plugin payload generators
  -pl-processor payload processors
  -w    wordlists to be used for payload

[output settings]
  -fmt  output file format(native, xml or json. only for file output) (default: native)
  -ie   ignore errors(will not output error message) (default: false)
  -ns   native stdout (default: false)
  -o    file to output
  -v    verbosity level(native output format only) (default: 1)

[recursion settings]
  -R    enable recursion mode(only support single fuzz keyword) (default: false)
  -rec-code     Recursion status code(http protocol only) (default: 200)
  -rec-depth    recursion depth(when recursion is enabled) (default: 2)
  -rec-regex    recursion when matched regex
  -rec-splitter splitter to be used to split recursion positions (default: /)

[error handle settings]
  -retry        max retries (default: 0)
  -retry-code   retry on status code(http protocol only)
  -retry-regex  retry when regex matched

[plugin settings]
  -preproc      preprocessor plugin to be used
  -react        reactor plugin to be used

SIMPLE USAGES:
fuzz URL:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com/FUZZ -w dict.txt::FUZZ

    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com/MILAOGIU -w dict.txt  # use default keyword

fuzz HTTP data:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com -w dict.txt::FUZZ -d "test=FUZZ"

use filters and matchers:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -w http://test.com/FUZZ -w dic.txt::FUZZ -mc 407 -fc 403-406 \
        -ms 123-154 -fs 10-100,120

use embedded payload processor to process payload:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com -w dict.txt::FUZZ -d "test=FUZZ" \
        -pl-processor suffix(".txt"),base64::FUZZ  # base64 encode

use embedded payload generators:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com/FUZZ \
        -pl-gen int(0,100,10)::FUZZ  # generate integer 0~100 with base 10

use multiple fuzz keywords and keyword process mode:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://FUZZ1/FUZZ2 -w dic1.txt::FUZZ1 \
        -w dic2.txt::FUZZ2  # default mode is "clusterbomb"

    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://FUZZ3/FUZZ4 -w dic3.txt::FUZZ3 \
        -w dic4.txt::FUZZ4 -mode pitchfork-cycle

refer to flag help information as above or https://github.com/nostalgist134/FuzzGIU/wiki for more usages

ADVANCED USAGES:
recursive jobs:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com/FUZZ -w dict.txt::FUZZ -R -rec-code 403 -rec-depth 4

use plugins:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com/?id=FUZZ \
        -pl-gen sqli::FUZZ  # will search ./plugins/payloadGenerators/sqli.(so/dll/dylib)

    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com -D "name=admin&pass=PASS" -w dict.txt::PASS \
        -pl-processor AES("1234567890abcdef")::PASS  # will search ./plugins/payloadProcessors/AES.(so/dll/dylib)

    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -w user.txt::NAME -w pass.txt::PASS \
        -u ssh://USER:PASS@test.com:22  # ./plugins/requestSenders/ssh.(so/dll/dylib)

    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com/FUZZ -w dict.txt::FUZZ \
        -preproc job_dispatch  # ./plugins/preprocessors/job_dispatch.(so/dll/dylib)

    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com/FUZZ -w dict.txt::FUZZ \
        -react fingerprint  # plugins/reactors/fingerprint.(so/dll/dylib)


        when fuzzGIU is executed without any args, it will init and create plugin directory "./plugin" to refer to plugins.
        there are 5 types of plugins can be used on current version: Preprocessor, PayloadGenerator, PayloadProcessor,
        RequestSender and Reactor. every plugin is of shared library format of current operating system, fuzzGIU will try to
        find plugin by plugin type and name at ./plugin/pluginType, make sure you put the plugin file to the right
        directory. you can find the usage of each type of plugin on https://github.com/nostalgist134/FuzzGIU/wiki. if you
        want to develop your own plugin, go check https://github.com/nostalgist134/FuzzGIUPluginKit, have fun :)
```

## 快速使用

以下是一些常见的使用示例

``````powershell
SIMPLE USAGES:
fuzz URL:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com/FUZZ -w dict.txt::FUZZ

    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com/MILAOGIU -w dict.txt  # use default keyword

fuzz HTTP data:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com -w dict.txt::FUZZ -d "test=FUZZ"

use filters and matchers:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -w http://test.com/FUZZ -w dic.txt::FUZZ -mc 407 -fc 403-406 \
        -ms 123-154 -fs 10-100,120 # match code 407, size 123~154; filter code 403~406, size 10~100,120

use embedded payload processor to process payload:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com -w dict.txt::FUZZ -d "test=FUZZ" \
        -pl-processor suffix(".txt"),base64::FUZZ  # base64 encode

use embedded payload generators:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com/FUZZ \
        -pl-gen int(0,100,10)::FUZZ  # generate integer 0~100 with base 10

use multiple fuzz keywords and keyword process mode:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://FUZZ1/FUZZ2 -w dic1.txt::FUZZ1 \
        -w dic2.txt::FUZZ2  # default mode is "clusterbomb"

    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://FUZZ3/FUZZ4 -w dic3.txt::FUZZ3 \
        -w dic4.txt::FUZZ4 -mode pitchfork-cycle
``````

### 窗口界面操作

通过上面的这些命令运行后，若未指定`-ns`选项，工具会打开一个termui窗口如下图所示。

![termui窗口](https://raw.githubusercontent.com/nostalgist134/FuzzGIU/main/imgs/fuzzGIU%20window.PNG)

在logo下方有4个小窗口，依次显示：**当前fuzz任务的全部信息**、**计数器**、**输出**以及**日志记录**。可以通过w/s键上下选中某个窗口，然后上/下键或者j/k键滑动查看窗口中的信息，选中的窗口边框会变蓝。按下p/r键可以暂停/恢复任务的执行，若任务暂停了，计数器的标题会改变。

单个任务执行完毕后以及所有任务执行完毕后，在日志窗口会有提示。所有任务执行完毕后程序不会自动退出，按下q键可以退出，在任务执行的过程中也可随时按下q键退出。

![fuzzGIU任务执行完毕](https://raw.githubusercontent.com/nostalgist134/FuzzGIU/main/imgs/fuzzGIU%20finish.PNG)

## 基础用法

### `-u`

`-u`用于指定fuzz的url，用法`-u scheme://url_body/path/`，工具自带的scheme有3种，`http/https`、`ws/wss`以及`dns`，第一种自不用多言，第二种是websocket协议，但是这个功能纯粹是我当时为了水过毕设加的，**没有经过测试，小心使用**。第三种`dns`协议并**不是用来对dns协议进行fuzz的，而是用来做域名枚举的，这个包我也没怎么测试过**。若工具检测到了不属于这3种的scheme，会**自动在插件目录中寻找RequestSender类型的插件，并使用插件来发送请求**，具体信息可在下文插件功能部分查看。

### `-w`

`-w`用来指定fuzz关键字所对应的字典，用法`-w dict.txt::FUZZ_KEYWORD`，字典列表与关键字之间使用`::`隔开。可以在命令行参数中指定多个`-w`参数，会自动识别，在单个`-w`参数中指定多个字典文件也是可行的，字典间需要通过逗号隔开，比如`-w dict1.txt,dict2.txt::FUZZ_KEYWORD`。但需要注意，此参数**在关键字未指定的情况下，会自动把指定的字典关联到默认关键字**`MILAOGIU`**上**。

### `-d`

`-d`参数用来指定请求体部分，但是这么说其实不太准确，因为请求体是属于http协议中的概念，而fuzzGIU为了对协议进行扩展，使用一个`Req`结构来表示请求：

``````go
HTTPSpec struct {
    Method     string   `json:"method,omitempty" xml:"method,omitempty"`
    Headers    []string `json:"headers,omitempty" xml:"header>headers,omitempty"`
    Version    string   `json:"version,omitempty" xml:"version,omitempty"`
    ForceHttps bool     `json:"force_https,omitempty" xml:"force_https,omitempty"`
}
// Req 请求对象
Req struct {
    URL      string   `json:"url,omitempty" xml:"url,omitempty"`
    HttpSpec HTTPSpec `json:"http_spec,omitempty" xml:"http_spec,omitempty"`
    Data     string   `json:"data,omitempty" xml:"data,omitempty"`
}
``````

`-d`参数指定的是**Req结构中的**`Data`**成员**，当然在**http协议fuzz中，这个成员就是拿来当成请求体用的**。

### `-r`

`-r`参数用来从文件中读取Req结构，其行为随文件的内容变化，具体如下：

1. 文件内容为http请求包，会自动识别并根据内容生成对应的Req请求结构。
2. 文件内容为json格式的Req请求对象，会将文件内容反序列化生成Req结构。
3. 文件内容不是以上两者，会将文件的内容全部填充到Req结构的`Data`成员中。

需要注意的是，**命令行参数也可指定这一结构中的成员值，且命令行参数指定的成员值优先级比**`-r`**参数所指定的更高，因此若在这一参数之外还有用别的命令行参数指定成员值，这里指定的值会被覆盖**。

### `-delay`

参数指定工具每次提交请求任务之后，以毫秒为单位的等待时间，防止请求速度过快导致触发可能的防御机制。

### `-t`

指定工具使用的协程池大小（并发数），默认值为64，可按需进行调整。理论上来讲值越大，执行任务的速度就越快，但是一台机子的资源是有限的，所以如果指定的值太大就没什么用了，而且协程太多反而可能消耗太多运算资源拖慢速度。

### `-timeout`

指定每个fuzz请求等待响应的最大时间（单位：秒）。注意**这个参数对于自定义的请求发送模块不是强制的**，工具只会把这一参数包装到请求发送相关的元信息中，如何处理这个参数取决于请求发送器插件的行为。

### `-f`(Filter)、`-m`(Matcher)系列参数

用来指定过滤条件与匹配条件相关的选项，这些选项会影响工具是否输出fuzz结果。目前对过滤和匹配条件支持以下几种

``````shell
# 过滤/匹配条件
-f(m)c   # 针对http返回状态码
-f(m)l   # 针对返回包的行数
-f(m)r   # 根据对返回包的正则匹配结果
-f(m)s   # 针对返回包的大小
-f(m)t   # 针对返回包的响应时间
-f(m)w   # 针对返回包的词数
``````

所有以**数字**作为单位的条件都**使用形如**`a-b,c,d-e,f,...`**的数字-横杠表达式**指定其**闭区间**范围，时间条件除外；返回包的正则匹配表达式为字符串，这个条件是**真的会按照正则表达式规则去匹配的**；时间条件使用单个时间（虽然没什么意义，因为基本不会遇到正好和单个时间匹配的包）或者**以逗号隔开的时间范围**来表示，时间条件的区间是**下闭上开**的，也就是说指定`-ft a,b`则对时间的匹配是在**a<=t<b**的范围内。
使用`-f(m)mode`参数指定当多个条件出现时，总条件的连接方式。支持两种方式：**or**与**and**。默认情况下过滤和匹配均使用`or`模式，即当多个过滤（匹配）条件出现时，只要有一个为真，就算过滤（匹配）。`and`模式则要求全部条件都为真才算过滤（匹配）。
工具遵循的输出规则为：**仅当结果视为不过滤（**`-f`**系列总条件不满足），且匹配条件（**`-m`**系列总条件满足），或者请求过程中发生错误时才会输出**。

### `-retry`系列参数

这一系列的参数用来指定工具是否会、在什么情况下会重试请求以及最多重试几次

``````shell
-retry        # 最大重试次数
-retry-code   # 遇到特定http状态码时重试
-retry-regex  # 遇到响应包匹配正则时重试
``````

当工具达到最大重试次数或者重试条件不满足时，工具会停止重试。

### http相关参数

和http fuzz相关的参数如下

```  shell
-H    # http头
-X    # http方法
-b    # cookies
-http2        # http2版本
-s    # 强制使用https
-x    # 代理
-F    # 跟随重定向
```

除`-x`外的其它参数会依次被填充到Req结构的HTTPSpec子结构中

``````go
HTTPSpec struct {
    Method     string	// -X
    Headers    []string // -H（允许多次指定）
    Version    string   // -http2
    ForceHttps bool     // -s
}
``````

`-x`参数会被存储到一个请求发送相关的元数据当中，工具发送请求时，无论是什么协议，总会向对应的请求发送模块传递这个元信息，因此`-x`**并不局限于http协议中**（只是由于设计失误，目前这个参数还是归类在此，详细信息可参考FuzzGIU wiki）。

### 输出设置

和输出设置相关的参数如下

``````shell
-fmt  # 输出格式，支持3种(native-原生输出格式、json、xml)，这个选项只对输出文件有用，屏幕输出是固定的native格式
-ie   # 忽略请求过程中出现的错误（ignore-error），不输出现错误的结果
-ns   # 使用原生stdout流（native-stdout）向屏幕输出结果（工具默认使用termui界面作为屏幕输出），固定以json格式输出
-o    # 输出文件名
-v    # 输出详细程度（只对输出格式为native起效，其它两种无论如何都输出全部信息），值从1~3
``````

### payload相关设置

``````shell
-mode # fuzz 关键字处理模式
-pl-gen       # payload 生成器
-pl-processor # payload 处理器
-w    # fuzz 字典
``````

和payload相关的设置总共有这4个，其中`-w`的详细用法上面已经介绍过了，这里不赘述。

`-mode`参数用来处理当请求中出现多个不同的fuzz关键字，或者单个关键字出现多次时，工具的处理模式。这个参数目前有4种值：`clusterbomb`、`pitchfork`、`pitchfork-cycle`和`sniper`。
前3种模式均用于处理出现多个不同关键字的场景，`clusterbomb`模式会枚举不同关键字对应的payload列表的所有组合（基本上和burp suite里面的同名模式是一样的）；`pitchfork`模式对每个关键字payload列表使用相同的下标，遍历到最短的payload列表结束为止；`pitchfork-cycle`模式则是`pitchfork`模式的改进版本，其迭代过程中每个关键字的列表下标仍然同步更新，但是较短的列表结束后，下标会从0再开始，循环往复，直到遍历到最长的列表结束。
`sniper`模式用于且仅能用于单个关键字在请求中出现多次的情况，这个模式下，工具会根据关键字出现的位置依次将特定位置的关键字替换为payload列表中的payload，并将其它位置的关键字替换为空。

`-pl-gen`参数用来指定fuzz关键字的payload生成器，注意，目前`-pl-gen`**与**`-w`**选项是互斥的，也就是说现在暂不支持同时使用payload生成器和字典来对某个关键字生成payload**。`-pl-gen`参数的使用方法与`-w`参数类似，使用`::`符号关联payload生成器列表和关键字。各个payload生成器间通过逗号隔开，单个payload生成器**伪函数调用表达式**来指定。

---

## 伪函数调用表达式的写法规则

1. 伪函数表达式的格式为：函数名([参数1, 参数2, 参数3, ...])。函数名即为使用的插件的名字。若参数列表为空，则括号也可省略
2. 参数支持4种类型：`int`、`float64`、`bool`、`string`
3. 字符串参数使用单引号或者双引号括起来，两种都是可接受的，但是必须配对
4. `bool`型参数使用全小写的`true`和`false`
5. `int`型参数支持两种进制，10进制和16进制，默认按照10进制数算，但是如果无法解析为10进制（含有字母），则尝试解析为16进制数；也可显式指定16进制数，在数字前面加上`0x`前缀即可

---

所有的插件均通过这个规则调用，`-h`显示的帮助信息的示例用法中有若干调用示例。

目前工具内置2种payload生成器，`int`和`permute`。
`int`生成器用来生成一个范围内的所有整数字符串。参数列表为`int(lower int, upper int, base int)`，`lower`参数为生成范围的下界（闭区间），`upper`为上界（开区间），`base`参数指定生成的数字的进制表示。
`permute`生成器用来生成一个字符串的所有排列。参数列表为`permute(s string, maxLen int)`，`s`参数为要排列的字符串，`maxLen`参数为生成排列列表的最大长度。

`-pl-processor`参数用于指定fuzz关键字的对应payload使用的处理器，同样使用伪函数调用表达式进行调用，使用`::`符号与关键字进行关联。对单个fuzz关键字也可指定多个处理器，会按照顺序依次调用，每个处理器处理后的payload会作为下一个处理器的输入。
内置的payload处理器有6种：`base64`、`urlencode`、`addslashes`、`stripslashes`、`suffix`和`repeat`。前3种处理器通过名字就可看出来做什么的，就不赘述。
`stripslashes`处理器会将payload开头的斜杆`/`去除，并且将payload中所有的2个以上连续的斜杆都换为单个斜杆。
`suffix`处理器会为payload添加后缀，接收一个`string`类型参数作为添加的后缀。
`repeat`处理器会简单地将payload重复多次，接收一个`int`类型参数表示重复的次数。

## 进阶用法

### 递归任务

递归模式是 FuzzGIU 用于深度探测目标的高级功能，适用于需要逐层挖掘资源的场景（如目录枚举、多级路径探测等）。通过`-R`启用后，工具会根据响应结果动态生成新的 fuzz 任务，实现自动化的深度探测。

#### 核心参数详解

- `-R`：启用递归模式（仅支持单个 fuzz 关键字，避免多关键字导致的逻辑冲突）。
- `-rec-code`：触发递归的 HTTP 状态码（默认 200），即当响应状态码匹配时，对结果进行二次 fuzz。
- `-rec-depth`：递归深度（默认 2），控制最大探测层级，防止无限递归消耗资源。
- `-rec-regex`：通过正则匹配响应内容触发递归，优先级高于`-rec-code`。
- `-rec-splitter`：用于分割递归位置的分隔符（默认`/`），适用于 URL 路径、目录等层级结构的拆分。

#### 使用场景与示例

**场景 1：多级目录递归枚举**
当探测目标网站的目录结构时，若一级目录（如`/admin/`）返回 200 状态码，可自动对其下的二级目录（如`/admin/user/`）进行探测，直到达到指定深度。

```powershell
# 递归探测http://test.com下的目录，遇到200状态码则继续深入，最大深度3
.\FuzzGIU.exe -u http://test.com/FUZZ -w dirs.txt::FUZZ -R -rec-code 200 -rec-depth 3
```

**场景 2：基于内容匹配的递归**
若目标响应中包含特定关键词（如 “更多路径”），可通过正则匹配触发递归，挖掘隐藏资源。

```powershell
# 当响应中包含"next_path:"时触发递归，分隔符使用默认的"/"
.\FuzzGIU.exe -u http://test.com/FUZZ -w initial_dirs.txt::FUZZ -R -rec-regex "next_path: (.*?)" -rec-depth 2
```

### 使用插件

FuzzGIU 通过插件系统扩展功能，支持自定义预处理、payload 生成、请求发送等逻辑，满足复杂场景的测试需求。插件为当前系统的共享库格式（`.so`/`.dll`/`.dylib`），需放置在`./plugin/[插件类型]/`目录下（如果这些目录没有创建，可不带任何参数运行一次工具，从而创建这些目录）。

#### 插件类型与作用

1. **Preprocessor（预处理插件）**
   在 fuzz 任务启动前对请求参数、字典等进行预处理（如动态生成字典、修改请求模板）。

   ```powershell
   # 使用job_dispatch预处理插件优化任务调度
   .\FuzzGIU.exe -u http://test.com/FUZZ -w big_dict.txt::FUZZ -preproc job_dispatch
   ```

2. **PayloadGenerator（payload 生成器插件）**
   替代字典（`-w`）动态生成 payload，适用于规则化、定制化的测试场景（如 SQL 注入、XSS 测试用例）。
   示例：`sqli`插件生成 SQL 注入测试 payload：

   ```powershell
   # 使用sqli插件生成注入payload，对ID参数进行fuzz
   .\FuzzGIU.exe -u http://test.com/?id=FUZZ -pl-gen sqli::FUZZ
   ```

3. **PayloadProcessor（payload 处理器插件）**
   对生成的 payload 进行二次处理（如加密、编码），满足目标系统的格式要求。
   示例：`AES`插件对密码 payload 进行 AES 加密：

   ```powershell
   # 对密码字典中的payload用AES加密后发送（密钥为1234567890abcdef）
   .\FuzzGIU.exe -u http://test.com/login -d "user=admin&pass=PASS" -w pass.txt::PASS -pl-processor AES("1234567890abcdef")::PASS
   ```

4. **RequestSender（请求发送插件）**
   扩展工具支持的协议（如 SSH、FTP、Redis 等），当 URL 的 scheme 不在默认支持的`http/https`、`ws/wss`、`dns`范围内时，工具会自动调用对应插件。
   示例：`ssh`插件用于 SSH 弱口令测试：

   ```powershell
   # 对SSH服务进行用户名/密码爆破（依赖ssh请求发送插件）
   .\FuzzGIU.exe -u ssh://USER:PASS@test.com:22 -w user.txt::USER -w pass.txt::PASS
   ```

5. **Reactor（响应处理器插件）**
   对响应结果进行自定义分析（如指纹识别、漏洞特征匹配），并输出结构化结果。
   示例：`fingerprint`插件识别目标服务器的中间件版本：

   ```powershell
   # 对响应进行指纹识别，输出服务器组件信息
   .\FuzzGIU.exe -u http://test.com/FUZZ -w endpoints.txt::FUZZ -react fingerprint
   ```

#### 插件开发与扩展

若内置组件无法满足需求，可基于[FuzzGIUPluginKit](https://github.com/nostalgist134/FuzzGIUPluginKit)开发自定义插件，遵循工具定义的接口规范，实现特定逻辑后编译为动态链接库即可使用。

# 特别感谢

特别致敬 [@ffuf](https://github.com/ffuf/ffuf) 项目，其理念与实现为本工具提供了重要启发，没有ffuf团队对于此类工具的探索，就没有这个项目。

特别感谢[@xch-space](https://github.com/xch-space)对项目命名提供的灵感。
