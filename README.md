虽然不知道会不会有人看，但是我还是写一个readme吧

# 项目介绍

FuzzGIU 是一款基于 Go 语言开发的web fuzzer，灵感来源于`ffuf`、`burp intruder`与`yakit web fuzzer`。适用于Web信息收集、漏洞扫描、API 测试等场景。

## 安装

安装fuzzGIU可以选择从源码编译

``````shell
git clone github.com/nostalgist134/FuzzGIU.git
cd FuzzGIU
go get
go build
``````

或者到release中下载编译好的可执行文件（暂时还没弄好，因为我没有mac系统，等我想出办法了再弄）

# 使用须知

本项目中所涉及的技术、思路和工具仅供以学习交流使用，任何人不得将其用于非法用途以及盈利等目的，否则后果自行承担。

目前使用http2协议会导致不定时的panic，我原本以为是我代码里面改http.transport导致的，结果我改了以后发现问题还是没办法解决，现在查不出来是什么问题，傻逼的要死net/http，之后想到办法再说

# 使用方法

执行 `FuzzGIU -h` 可查看完整的命令行帮助信息：

```powershell
PS H:\tools\fuzz\FuzzGIU> .\FuzzGIU.exe -h
Usage of H:\tools\fuzz\FuzzGIU\FuzzGIU.exe:
        H:\tools\fuzz\FuzzGIU\FuzzGIU.exe [options]
options are shown below. when fuzzGIU is executed without any args,
it will init and create plugin directory

GENERAL OPTIONS:
  -d    request data
  -delay        delay between each job submission(millisecond) (default: 0)
  -r    request file
  -t    routine pool size (default: 64)
  -timeout      timeout(second) (default: 10)
  -u    url to giu

MATCHER OPTIONS:
  -mc   match status code from response (default: 200,204,301,302,307,401,403,405,500)
  -ml   match amount of lines in response
  -mmode        matcher set operator (default: or)
  -mr   match regexp
  -ms   match response size
  -mt   match time(millisecond) to the first response byte
  -mw   match amount of words in response

FILTER OPTIONS:
  -fc   filter status code from response
  -fl   filter amount of lines in response
  -fmode        filter set operator (default: or)
  -fr   filter regexp
  -fs   filter response size
  -ft   filter time(millisecond) to the first response byte
  -fw   filter amount of words in response

REQUEST OPTIONS:
  -F    follow redirects (default: false)
  -H    request headers to be used
  -X    request method (default: GET)
  -b    Cookies
  -http2        force http2 (default: false)
  -ra   http random agent (default: false)
  -s    force https (default: false)
  -x    proxies

PAYLOAD OPTIONS:
  -mode mode for keywords used, basically the same as those in burp suite (default: clusterbomb)
  -pl-gen       plugin payload generators
  -pl-processor payload processors
  -w    wordlists to be used for payload

OUTPUT OPTIONS:
  -fmt  output file format(native, xml or json. only for file output) (default: native)
  -ie   ignore errors(will not output error message) (default: false)
  -ns   native stdout (default: false)
  -o    file to output
  -v    verbosity level(native output format only) (default: 1)

RECURSION OPTIONS:
  -R    enable recursion mode(only support single fuzz keyword) (default: false)
  -rec-code     Recursion status code(request protocol only)
  -rec-depth    recursion depth(when recursion is enabled) (default: 2)
  -rec-regex    recursion when matched regex
  -rec-splitter splitter to be used to split recursion positions (default: /)

ERROR HANDLE OPTIONS:
  -retry        max retries (default: 0)
  -retry-code   retry on status code(request protocol only)
  -retry-regex  retry when regex matched

PLUGIN OPTIONS:
  -preproc      preprocessor plugin to be used
  -react        reactor plugin to be used

SIMPLE USAGES:
fuzz URL:
    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com/FUZZ -w dict.txt::FUZZ

    H:\tools\fuzz\FuzzGIU\FuzzGIU.exe -u http://test.com/MILAOGIU -w dict.txt  # use default keyword

fuzz Request data:
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
```

## 快速使用

- **URL 路径 Fuzz:**

  ```powershell
  # 指定关键字
  .\FuzzGIU.exe -u http://test.com/FUZZ -w directory_list.txt::FUZZ
  # 使用默认关键字 "MILAOGIU"
  .\FuzzGIU.exe -u http://test.com/MILAOGIU -w directory_list.txt
  ```

- **HTTP 请求体 Fuzz:**

  ```powershell
  .\FuzzGIU.exe -u http://test.com/login -d username=admin&password=FUZZ -w passwords.txt::FUZZ
  ```

- **使用匹配器 (Matcher) 和过滤器 (Filter):**

  ```powershell
  # 匹配状态码200或301，且响应大小在1000-2000字节之间；过滤状态码403或404，以及响应大小在0-500字节或5000字节的结果
  .\FuzzGIU.exe -u http://test.com/FUZZ -w fuzz_dict.txt::FUZZ -mc 200,301 -fc 403,404 -ms 1000-2000 -fs 0-500,5000
  ```

- **使用内置 Payload 处理器:**

  ```powershell
  # 对每个 payload 先添加 '.bak' 后缀，再进行 Base64 编码。
  .\FuzzGIU.exe -u http://test.com/download -w filenames.txt::FUZZ -d file=FUZZ -pl-processor suffix('.bak'),base64::FUZZ
  ```

- **使用内置 Payload 生成器:**

  ```powershell
  # 生成数字 1 到 99 作为 FUZZ 关键字使用的payload列表
  .\FuzzGIU.exe -u http://test.com/user?id=FUZZ -pl-gen int(1,100)::FUZZ
  ```

- **指定关键字处理模式:**

  ```powershell
  # Clusterbomb 模式 (默认): 遍历所有组合
  .\FuzzGIU.exe -u http://FUZZHOST/FUZZPATH -w subdomains.txt::FUZZHOST -w paths.txt::FUZZPATH
  
  # Pitchfork-Cycle 模式: 列表循环对齐
  .\FuzzGIU.exe -u http://FUZZUSER:FUZZPASS@test.com -w usernames.txt::FUZZUSER -w passwords.txt::FUZZPASS -mode pitchfork-cycle
  
  # sniper 模式: 接收单个关键字，根据其出现在请求中的位置依次替换为payload，其它位置替换为空
  .\fuzzGIU.exe -u http://test.com/FUZZ -d user=FUZZ -H Header: FUZZ -w dic.txt::FUZZ -mode sniper
  ```

### 窗口界面操作

通过命令运行fuzzGIU后，若未指定`-ns`选项，工具会打开一个如下图所示的窗口

![termui窗口](https://raw.githubusercontent.com/nostalgist134/FuzzGIU/main/imgs/fuzzGIU%20window.PNG)

在logo下方有4个窗口，依次显示如下内容：

- **GLOBAL_INFORMATION**:  当前fuzz任务的信息

- **PROGRESS**: 计数器，包含当前任务的进度、总进度、请求发送速率以及消耗时间

- **OUTPUT**: 输出窗口，符合输出条件的结果会在此处输出

- **LOGS**: 日志窗口

**界面操作：**

- **`W`/`S` 键**：上下移动焦点，选中窗口边框变蓝。
- **`↑`/`↓` 或 `J`/`K` 键**：在选中窗口内滚动内容。
- **`L` 键**：锁定/取消锁定输出窗口，默认情况下，输出窗口总会聚焦在最后一个输出上，按下此键可取消锁定，从而自由移动。
- **`P`/`R` 键**：暂停/恢复任务执行（计数器标题会变化）。
- **`Q` 键**：随时退出程序。

单个任务执行完毕后以及所有任务执行完毕后，在日志窗口会有提示。所有任务执行完毕后程序不会自动退出，按下q键可以退出，在任务执行的过程中也可随时按下q键退出。

![fuzzGIU任务执行完毕](https://raw.githubusercontent.com/nostalgist134/FuzzGIU/main/imgs/fuzzGIU%20finish.PNG)

## 基础用法

### `-u`

`-u`用于指定fuzz的url，工具自带的scheme有3种：`http/https`、`ws/wss`以及`dns`。第一种自不用多言，第二种是websocket协议，但是这个功能纯粹是我当时为了水过毕设加的，**没有经过测试，小心使用**。第三种`dns`协议并**不是用来对dns协议进行fuzz的，而是用来做域名枚举的，这个包我也没怎么测试过**。若工具检测到了这3种之外的scheme，会自动在插件目录中寻找RequestSender类型的插件，并使用插件来发送请求，具体信息可在下文[使用插件](#使用插件)部分查看。

`-u`参数指定的url中可包含fuzz关键字，在fuzz过程中会自动被替换。

### `-w`

`-w`用来指定fuzz关键字所对应的字典，用法为`-w dict.txt::FUZZ_KEYWORD`，字典列表通过`::`符号与fuzz关键字相关联。可以在命令行参数中指定多个`-w`参数，会自动识别，在单个`-w`参数中指定多个字典文件也是可行的，字典间需要通过逗号隔开，比如`-w dict1.txt,dict2.txt::FUZZ_KEYWORD`。也可省略关键字部分，此时工具会将字典关联到默认关键字`MILAOGIU`上。

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

**注意**：其它命令行参数也可指定Req结构中的成员值，且指定优先级比本参数更高，会覆盖此命令指定的值。

### `-delay`

参数指定工具每次提交请求任务之后，以毫秒为单位的等待时间，防止请求速度过快导致触发可能的防御机制或者资源占用过多。

### `-t`

指定工具使用的协程池大小（并发数），默认值为64，可按需进行调整。理论上来讲值越大，执行任务的速度就越快，但是一台机子的资源是有限的，所以如果指定的值太大就没什么用了，而且协程太多反而可能导致资源竞争。

### `-timeout`

指定每个fuzz请求等待响应的最大时间（单位：秒）。**注意：** 对于自定义请求发送插件 (`RequestSender`)，工具只会把这一参数包装到请求发送相关的元信息中，是否遵守此超时取决于插件实现。

### `-f`(Filter)、`-m`(Matcher)系列参数

用来指定过滤条件与匹配条件相关的选项，这些选项会影响工具是否输出fuzz结果。目前对过滤和匹配条件支持以下几种

``````shell
# 过滤/匹配条件
-f(m)c   # 针对http返回状态码
-f(m)l   # 针对返回包的行数
-f(m)r   # 根据对返回包的正则表达式匹配结果
-f(m)s   # 针对返回包的大小
-f(m)t   # 针对返回包的响应时间
-f(m)w   # 针对返回包的词数
``````

所有以**数字**作为单位的条件都**使用形如**`a-b,c,d-e,f,...`**的数字-横杠表达式**指定其**闭区间**范围，时间条件除外。时间条件使用单个时间（虽然没什么意义，因为基本不会遇到正好和单个时间匹配的包）或者**以逗号隔开的时间范围**来表示，时间条件的区间是**下闭上开的毫秒区间**，指定`-ft a,b`对应的范围为`[a,b)`。

**逻辑模式 (`-mmode`/`-fmode`)**

- `or` (默认)：任意一个条件满足即视为匹配/过滤。
- `and`：所有条件都必须满足才视为匹配/过滤。

**输出规则：** 仅当结果**未通过过滤器** (`-f*` 总条件不满足) **且** **通过匹配器** (`-m*` 总条件满足) **或** 请求过程中发生错误时，结果才会被输出。

### `-retry`系列参数

这一系列的参数用来指定工具是否会、在什么情况下会重试请求以及最多重试几次

``````shell
-retry        # 最大重试次数
-retry-code   # 遇到特定http状态码时重试
-retry-regex  # 遇到响应包匹配正则时重试
``````

当工具达到最大重试次数或者重试条件不满足时，工具会停止重试。

### 请求设置

和请求设置相关的参数如下

```  shell
-H    # http头
-X    # http方法
-b    # cookies
-http2        # 使用http2
-s    # 强制使用https
-x    # 代理
-F    # 跟随重定向
-ra   # 随机ua头
```

除`-x`、`-F`和`-ra`外的其它参数会依次被填充到Req结构的HTTPSpec子结构中

``````go
HTTPSpec struct {
    Method     string	// -X
    Headers    []string // -H（允许多次指定）
    Version    string   // -http2
    ForceHttps bool     // -s
}
``````

且这些参数（除`-s`外）都可以包含fuzz关键字，fuzz过程中会自动识别并替换。

`-x`、`-F`参数会被存储到一个请求发送相关的元数据当中，工具发送请求时，无论是什么协议，总会向对应的请求发送模块传递这个元信息，但若使用的协议是插件，则怎么处理代理取决于协议本身的逻辑。目前内置的协议仅支持http代理，不支持socks系列。

`ra`头通过设置一个全局变量来启用随机ua头。

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

和payload相关的设置总共有这4个，`-w`参数的详细用法上面已经介绍过了，这里不赘述。

`-mode`参数用来处理当请求中出现多个不同的fuzz关键字，或者单个关键字出现多次时，工具的处理模式。这个参数目前有4种值：`clusterbomb`、`pitchfork`、`pitchfork-cycle`和`sniper`。

前3种模式均用于处理出现多个不同关键字的场景，`clusterbomb`模式会枚举不同关键字对应的payload列表的所有组合（基本上和burp suite里面的同名模式是一样的）；`pitchfork`模式对每个关键字payload列表使用相同的下标，遍历到最短的payload列表结束；`pitchfork-cycle`模式则是`pitchfork`模式的改进版本，其迭代过程中每个关键字的列表下标仍然同步更新，但是较短的列表结束后，下标会从0再开始，循环往复，直到最长的列表结束。

`sniper`模式用于且仅能用于单个关键字在请求中出现多次的情况：工具会根据关键字出现的位置依次将特定位置的关键字替换为payload列表中的payload，并将其它位置的关键字替换为空。

`-pl-gen`参数用来指定fuzz关键字的payload生成器，注意，目前`-pl-gen`**与**`-w`**选项是互斥的，暂不支持同时使用payload生成器和字典来对某个关键字生成payload**。`-pl-gen`参数的使用方法与`-w`参数类似，使用`::`符号关联payload生成器列表和关键字。各个payload生成器间通过逗号隔开，单个payload生成器 [伪函数调用表达式](#插件调用) 来指定。

目前工具内置2种payload生成器：

- `int(lower, upper, base)`: 生成 `[lower, upper)` 范围内指定 `base` 进制 (通常为 10) 的数字字符串。e.g., `int(1, 100, 10)` 生成 "1" 到 "99"，`base`参数可省略，默认为10。
- `permute(s, maxLen)`: 生成字符串 `s` 的所有排列组合，最多 `maxLen` 个结果。e.g., `permute("abc", 10)`，若`maxLen`省略或为-1，则不限制。

`-pl-processor`参数用于指定fuzz关键字的对应payload使用的处理器，同样使用 [伪函数调用表达式](#插件调用) 进行调用，使用`::`符号与关键字进行关联。对单个fuzz关键字也可指定多个处理器，会按照顺序依次调用，每个处理器处理后的payload会作为下一个处理器的输入。
内置的6种payload处理器如下：

- `base64`: Base64 编码 payload。
- `urlencode`: URL 编码 payload。
- `addslashes`: 添加反斜杠转义特殊字符 (如 `'` -> `\'`)。
- `stripslashes`: 去除开头的 `/` 并将连续多个 `/` 替换为单个 `/`。
- `suffix(s)`: 给 payload 添加后缀 `s`。e.g., `suffix(".php")`。
- `repeat(n)`: 将 payload 重复 `n` 次。e.g., `repeat(3)` 将 "a" 变为 "aaa"。

## 进阶用法

### 递归任务

递归模式是 FuzzGIU 用于深度探测目标的高级功能，适用于需要逐层挖掘资源的场景（如目录枚举、多级路径探测等）。通过`-R`启用后，工具会根据响应结果动态生成新的 fuzz 任务，实现自动化的深度探测。

#### 核心参数详解

- `-R`：启用递归模式（仅支持单个 fuzz 关键字，避免多关键字导致的逻辑冲突）。
- `-rec-code`：触发递归的 HTTP 状态码（默认 200），即当响应状态码匹配时，对结果进行二次 fuzz。
- `-rec-depth`：递归深度（默认 2），控制最大探测层级，防止无限递归消耗资源。
- `-rec-regex`：通过正则匹配响应内容触发递归。
- `-rec-splitter`：用于分割递归位置的分隔符（默认`/`），适用于 URL 路径、目录等层级结构的拆分。

递归条件`-rec-code`与`-rec-regex`同时指定的情况下使用`or`连接。

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

FuzzGIU 通过插件系统扩展功能，支持自定义预处理、payload 生成、请求发送等逻辑，满足复杂场景的测试需求。插件为当前系统的共享库格式（`.so`/`.dll`/`.dylib`），需放置在`./plugins/[插件类型]/`目录下（若尚未创建插件目录，可不带任何参数运行一次工具，从而创建这些目录）。

```powershell
PS H:\tools\fuzz\FuzzGIU> .\FuzzGIU.exe
Checking/initializing environment...
Checking directory ./plugins/......Created.
Checking directory ./plugins/payloadGenerators/......Created.
Checking directory ./plugins/payloadProcessors/......Created.
Checking directory ./plugins/preprocessors/......Created.
Checking directory ./plugins/requestSenders/......Created.
Checking directory ./plugins/reactors/......Created.
Done.
For help, use -h flag
```

#### 插件调用

fuzzGIU的插件调用遵循以下的被称为伪函数调用表达式的规则：

1. 伪函数表达式的格式为：`fn_0([arg0, arg1, arg2, ...]),fn_1([arg0, arg1, ...]),...`。函数名`fn_n`即为使用的插件的名字，不同的函数间通过`,`隔开。若某个函数的参数列表为空，则括号也可省略
2. 参数支持4种类型：`int`、`float64`、`bool`、`string`
3. 字符串参数使用单引号或者双引号括起来，两种都是可接受的，但是必须配对
4. `bool`型参数使用全小写的`true`和`false`
5. `int`型参数支持两种进制，10进制和16进制，默认按照10进制数算，但是如果无法解析为10进制（含有字母），则尝试解析为16进制数；也可显式指定16进制数，在数字前面加上`0x`前缀即可
6. 默认参数：部分插件调用时传入默认参数，这些参数可能是fuzz过程中使用的**结构体**或者**数据**，调用语句中无需指定，工具会根据上下文自动处理，不同插件的默认参数参考下文 [插件类型与作用](#插件类型与作用) 章节。各个fuzz结构体的细节与作用可参考wiki中的 [相关章节](https://github.com/nostalgist134/FuzzGIU/wiki/FuzzGIU部分实现细节#模糊测试相关结构体)
7. 自定义参数：在默认参数之外，插件开发者可以在插件中指定任意不违反操作系统限制的数量的自定义参数，扩展插件的功能，这些参数实际上就是通过规则1中的参数列表传入

无论是内置的组件还是自定义的插件都根据这些规则来进行带参调用。工具会将函数名作为插件名，到`./plugins/[插件类型]`目录下找到对应名字的动态链接库并调用。

#### 插件类型与作用

1. **Preprocessor（预处理插件）**

   + **作用**：在 fuzz 任务启动前对请求参数、字典等进行预处理（如动态生成字典、修改请求模板）

   + **默认参数**：`*fuzzTypes.Fuzz`->当前使用的fuzz任务结构体
   + **返回值**：`*fuzzTypes.Fuzz`->处理后的fuzz任务

   + **复合行为**：若指定了多个预处理插件，在插件链上每一个插件的返回任务都会作为默认参数传递给下一个插件。

   ```powershell
   # 使用job_dispatch预处理插件优化任务调度
   .\FuzzGIU.exe -u http://test.com/FUZZ -w big_dict.txt::FUZZ -preproc job_dispatch
   ```

2. **PayloadGenerator（payload 生成器插件）**

   + **作用**：替代字典（`-w`）动态生成 payload，适用于规则化、定制化的测试场景（如 SQL 注入、XSS 测试用例）。
   + **默认参数**：无
   + **返回值**：`[]string`切片->生成的payload
   + **复合行为**：可对单个关键字指定多个payload生成器，每个生成器生成的payload都会添加到总列表中

   示例：`sqli`插件生成 SQL 注入测试 payload：

   ```powershell
   # 使用sqli插件生成注入payload，对ID参数进行fuzz
   .\FuzzGIU.exe -u http://test.com/?id=FUZZ -pl-gen sqli::FUZZ
   ```

3. **PayloadProcessor（payload 处理器插件）**

   + **作用**：对生成的 payload 进行二次处理（如加密、编码），满足目标系统的格式要求。
   + **默认参数**：`string`->要处理的payload，从关键字对应的payload列表中取出
   + **返回值**：`string`->处理后的payload
   + **复合行为**：若指定了多个payload处理器，则插件链上每个插件返回的处理后的payload都会作为默认参数传递给下一插件

   示例：`AES`插件对密码 payload 进行 AES 加密：

   ```powershell
   # 对密码字典中的payload先用AES加密（密钥为1234567890abcdef），再用base64编码
   .\FuzzGIU.exe -u http://test.com/login -d "user=admin&pass=PASS" -w pass.txt::PASS -pl-processor AES("1234567890abcdef"),base64::PASS
   ```

4. **RequestSender（请求发送插件）**

   + **作用**：扩展工具支持的协议
   + **默认参数**：`*fuzzTypes.SendMeta`->一个包含了请求本身和请求相关设置的上下文结构
   + **返回值**：`*fuzzTypes.Resp`->发送请求后接收到的响应结构
   + **复合行为**：本插件不支持复合调用
   + **其它注意事项**：这类插件通过`-u`指定的url的scheme字段隐式调用，当`-u`指定的scheme不在工具预置支持的协议范围内，工具就会根据其scheme在`./plugins/requestSenders`目录中寻找对应名字的动态链接库。由于调用过程中不涉及伪函数表达式，因此此类插件无法接收自定义参数。

   示例：`ssh`插件用于 SSH 弱口令测试：

   ```powershell
   # 对SSH服务进行用户名/密码爆破（依赖ssh请求发送插件）
   .\FuzzGIU.exe -u ssh://USER:PASS@test.com:22 -w user.txt::USER -w pass.txt::PASS
   ```

5. **Reactor（响应处理器插件）**

   + **作用**：对请求和响应结果进行综合性的自定义分析（如指纹识别、漏洞特征匹配），并输出结构化结果。
   + **默认参数**：`*fuzzTypes.Req`->请求结构体、`*fuzzTypes.Resp`->响应结构体
   + **返回值**：`*fuzzTypes.Reaction`->结构化的响应结果
   + **复合行为**：本插件不支持复合调用

   示例：`fingerprint`插件识别目标服务器的中间件版本：

   ```powershell
   # 对响应进行指纹识别，输出服务器组件信息
   .\FuzzGIU.exe -u http://test.com/FUZZ -w endpoints.txt::FUZZ -react fingerprint
   ```

#### 插件开发与扩展

若内置组件无法满足需求，可基于 [FuzzGIUPluginKit](https://github.com/nostalgist134/FuzzGIUPluginKit) 与go编译器开发自定义插件，遵循工具定义的接口规范，实现特定逻辑后编译为动态链接库即可使用。

# 特别感谢

## 项目

+ [ffuf/ffuf: Fast web fuzzer written in Go](https://github.com/ffuf/ffuf)
+ [yaklang/yakit: Cyber Security ALL-IN-ONE Platform](https://github.com/yaklang/yakit)
+ [Burp Suite - Application Security Testing Software - PortSwigger](https://portswigger.net/burp)

没有这些项目团队对于本类工具的探索与启发，就没有当前项目。

## 个人

特别感谢[@xch-space](https://github.com/xch-space)对项目命名提供的灵感。
