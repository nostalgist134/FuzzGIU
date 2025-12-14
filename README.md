# 项目介绍

FuzzGIU是个基于golang开发的简单的小玩具，用来做web fuzz。

## 安装

安装FuzzGIU可以选择从源码编译（建议go版本在1.25以上）

```shell
git clone github.com/nostalgist134/FuzzGIU.git
cd FuzzGIU
go get
go build
```

或者到release页面中下载编译的可执行文件

# 注意事项

本项目中所涉及的技术、思路和工具仅供以学习交流使用，任何人不得将其用于非法用途以及盈利等目的，否则后果自行承担。

这个项目不支持在32位系统上使用。

命名`FuzzGIU`是刻意为之，不是拼写错误。

# 使用方法

## 快速使用

- **URL 路径 Fuzz（目录扫描）:**

  ```powershell
  # 指定占位符
  .\FuzzGIU.exe -u http://test.com/FUZZ -w directory_list.txt::FUZZ
  
  # 使用默认占位符 "MILAOGIU"
  .\FuzzGIU.exe -u http://test.com/MILAOGIU -w directory_list.txt
  ```

- **HTTP 请求体/方法 Fuzz:**

  ```powershell
  # -d选项fuzz请求体
  .\FuzzGIU.exe -u http://test.com/login -d username=admin&password=FUZZ -w passwords.txt::FUZZ
  
  # -X选项fuzz/指定http请求方法
  .\FuzzGIU.exe -u http://test.com/ -X METHOD -w methods.txt::METHOD
  ```

- **使用匹配器 (Matcher) 和过滤器 (Filter):**

  ```powershell
  # 匹配状态码200或301，且响应大小在1000-2000字节之间；过滤状态码403或404，以及响应大小在0-500字节或5000字节的结果
  .\FuzzGIU.exe -u http://test.com/FUZZ -w fuzz_dict.txt::FUZZ -mc 200,301 -fc 403,404 -ms 1000-2000 -fs 0-500,5000
  ```

- **使用内置 Payload 处理器:**

  ```powershell
  # 对每个 payload 先添加 '.bak' 后缀，再进行 Base64 编码。
  .\FuzzGIU.exe -u http://test.com/download -w filenames.txt::FUZZ -d file=FUZZ -pl-proc suffix('.bak'),base64::FUZZ
  ```

- **使用内置 Payload 生成器:**

  ```powershell
  # 生成数字1~99作为FUZZ占位符使用的payload列表
  .\FuzzGIU.exe -u http://test.com/user?id=FUZZ -pl-gen int(1,100)::FUZZ
  ```

- **指定迭代模式:**

  ```powershell
  # clusterbomb 模式 (默认): 遍历所有组合
  .\FuzzGIU.exe -u http://FUZZHOST/FUZZPATH -w subdomains.txt::FUZZHOST -w paths.txt::FUZZPATH
  
  # pitchfork/pitchfork-cycle模式: 列表循环对齐
  .\FuzzGIU.exe -u http://FUZZUSER:FUZZPASS@test.com -w usernames.txt::FUZZUSER -w passwords.txt::FUZZPASS -iter pitchfork/pitchfork-cycle
  
  # sniper 模式: 单个占位符在请求中出现多次，根据其出现的位置依次替换为payload，其它位置替换为空
  .\FuzzGIU.exe -u http://test.com/FUZZ -d user=FUZZ -H "Header: FUZZ" -w dic.txt::FUZZ -iter sniper
  ```

## 帮助信息与环境初始化

执行 `FuzzGIU -h` 即可查看完整的命令行帮助信息；不带参数地执行`FuzzGIU`会使其进行环境初始化，创建插件目录。

## 命令行参数

### `-u`

`-u`用于指定请求url，工具通过url的scheme字段决定使用什么方式发送请求。自带的scheme有2种：`http/https`与`ws/wss`。第一种自不用多言，第二种是websocket协议，但是这个功能纯粹是我当时为了水过毕设加的，**没有经过测试，小心使用**。

若工具检测到了这2种之外的scheme，会自动在插件目录中寻找Requester类型的插件，并使用插件来发送请求，具体信息可在下文[使用插件](#使用插件)部分查看。

若url中不包含scheme字段，则采用http协议。

url任意部分的fuzz占位符都会被替换。

### `-w`

`-w`指定fuzz占位符对应的字典，用法为`-w dict.txt::FUZZ_KEYWORD`，通过`::`符号与fuzz占位符绑定。

`-w`参数允许在命令行中多次出现，比如`-w dict1.txt::USER -w dict2.txt::PASS`，可为多个占位符指定字典，或为单个占位符指定多个字典。

单个`-w`中亦可指定多个字典文件，其间通过逗号隔开，比如`-w dict1.txt,dict2.txt::FUZZ_KEYWORD`。多个字典文件通过逗号连接时，其内容会被**顺序合并**为一个 payload 列表，等效于 `cat dict1.txt dict2.txt > merged.txt` 后使用 `-w merged.txt::KEYWORD`。

**注意**：

+ 占位符部分可省略，这种情况下字典将关联到默认占位符`MILAOGIU`上。
+ `-pl-gen`命令行参数通过payload生成器插件生成payload，若某个占位符已经绑定了`-pl-gen`参数，则不能再为其绑定字典。

### `-d`与`-df`

`-d`/`-df`参数用来指定请求体部分，但是这么说其实不太准确，因为请求体是属于http协议中的概念，而FuzzGIU为了对协议进行扩展，使用一个`Req`结构来表示请求：

```go
type HTTPSpec struct {
    Method      string   `json:"method,omitempty" xml:"method,omitempty"`
    Headers     []string `json:"headers,omitempty" xml:"header>headers,omitempty"`
    Proto       string   `json:"proto,omitempty" xml:"proto,omitempty"`
    ForceHttps  bool     `json:"force_https,omitempty" xml:"force_https,omitempty"`
    RandomAgent bool     `json:"http_random_agent,omitempty"`
}
type Req struct {
    URL      string   `json:"url,omitempty" xml:"url,omitempty"`
    HttpSpec HTTPSpec `json:"http_spec,omitempty" xml:"http_spec,omitempty"`
    Fields   []Field  `json:"fields,omitempty" xml:"fields,omitempty"`
    Data     []byte   `json:"data,omitempty" xml:"data,omitempty"`
}
```

`-d`参数指定的是**Req结构中的**`Data`**成员**，当然在**http协议fuzz中，这个成员就是拿来当成请求体用的**。

`-df`：从文件中读取`Data`。

**注意**：工具不会自动根据请求体设置`Content-Type`头。

### `-r`

`-r`参数用来从文件中读取Req结构并使用，其行为随文件的内容变化，具体如下：

1. 文件内容为http请求包，会自动识别并根据内容生成对应的Req请求结构，当Proto字段恰好为`HTTP/2`时请求会采用http/2协议。
2. 文件内容为json格式的请求对象，会将文件内容反序列化生成Req结构。
3. 文件内容不是以上两者，会将文件的内容全部填充到Req结构的`Data`成员中，类似于`-d`的作用。

**注意**：其它命令行参数指定的优先级比本参数更高，会覆盖此命令指定的值。

### `-delay`

`-delay`参数指定工具每次提交请求任务之后的等待时间，用于控制请求速度，参数类型为字符串类型，格式为数字+单位（举例：`1.5ms`、`200us`、`3.75s`等）。

### `-t`与`-c`

`-t`用于指定单任务的并发数，默认为64，可按需进行调整；`-c`用于指定最大并发执行的任务数，默认为5。

### `-timeout`

指定每个fuzz请求等待响应的最大时间（单位：秒）。**注意：** 对于自定义请求发送插件，工具只会把这一参数包装到请求发送相关的元信息中，是否遵守此超时取决于插件实现。

### `-f`(Filter)、`-m`(Matcher)系列参数

用来指定过滤条件与匹配条件相关的选项，这些选项会影响工具是否输出fuzz结果。目前对过滤和匹配条件支持以下几种

```shell
# 过滤/匹配条件
-f(m)c   # 针对http返回状态码
-f(m)l   # 针对返回包的行数
-f(m)r   # 根据对返回包的正则表达式匹配结果
-f(m)s   # 针对返回包的大小
-f(m)t   # 针对返回包的响应时间
-f(m)w   # 针对返回包的词数
```

所有以**数字**作为单位的条件都**使用形如**`a-b,c,d-e,f,...`**的数字-横杠表达式**指定其**闭区间**范围，时间条件除外。时间条件使用单个时间（虽然没什么意义，因为基本不会遇到正好和单个时间匹配的包）或者**以逗号隔开的时间范围**来表示，时间条件的区间是**下闭上开的毫秒区间（`-ft a,b`->`a<=t<b`）**。

**逻辑模式 (`-mmode`/`-fmode`)**

- `or` (默认)：任意一个条件满足即视为匹配/过滤。
- `and`：所有条件都必须满足才视为匹配/过滤。

**输出规则：** 仅当结果**未通过过滤器** (`-f*` 聚合条件不满足) **且** **通过匹配器** (`-m*` 聚合条件满足) **或** 请求过程中发生错误时，结果才会被输出。

### `-retry`系列参数

这一系列的参数用来指定工具是否会、在什么情况下会重试请求以及最多重试几次

```shell
-retry        # 最大重试次数
-retry-code   # 遇到特定http状态码时重试
-retry-regex  # 遇到响应包匹配正则时重试
```

达到最大重试次数或者重试条件不满足时，工具会停止重试。`-retry-code`与`-retry-regex`使用`or`逻辑连接。

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
-r	  # 上面已经介绍过了
```

除`-x`、`-F`和`-ra`外的其它参数会依次被填充到Req结构的HTTPSpec子结构中

```go
type HTTPSpec struct {
    Method      string   `json:"method,omitempty" xml:"method,omitempty"`
    Headers     []string `json:"headers,omitempty" xml:"header>headers,omitempty"`
    Proto       string   `json:"proto,omitempty" xml:"proto,omitempty"`
    ForceHttps  bool     `json:"force_https,omitempty" xml:"force_https,omitempty"`
    RandomAgent bool     `json:"http_random_agent,omitempty"`
}
```

且这些参数（除`-s`外）都可以包含fuzz占位符，fuzz过程中会自动识别并替换。

`-x`、`-F`参数会被存储到一个请求发送相关的元数据当中，工具内置的协议能够正常处理这些参数，但若使用插件协议，则取决于插件内部实现。

`-ra`启用随机`User-Agent`头，对应的是`RandomAgent`成员。在http fuzz中，如果未指定ua头，也未指定`-ra`选项，使用的默认ua头为`milaogiu browser (21.1)`。

**注意**：

+ 由于`net/http`库本身的缺陷，若指定了`-http2`选项，内存占用会飙升至少10倍，并且请求会变得极不稳定，虽然工具提供了此选项，但是不建议使用。
+ 目前版本中的内置协议仅支持http代理。

### 输出设置

和输出设置相关的参数如下

```shell
-fmt  # 输出格式（目前支持native、json与xml这3种格式）
-ie   		# 忽略请求过程中出现的错误（ignore-error），不输出现错误的结果
-ns   		# 直接输出结果到stdout（工具默认使用tview界面作为屏幕输出），固定以json格式输出
-tview 		# 是否启用tview窗口输出（默认开启，引入 -tview=false 来关闭tview窗口输出）
-out-file   # 输出文件名
-out-url	# 指定将输出post到http url
-v    		# 输出详细程度（只对输出格式为native起效，其它两种无论如何都输出全部信息），范围1~3
```

**注意**：

+ tview输出窗口忽略`-fmt`参数，总是采用`native`格式输出。
+ 在普通模式下启动时，工具默认使用tview窗口界面，如果不希望使用此界面需手动设置`-tview=false`或者启用`-ns`选项；以http api模式启动时，不允许提交使用tview窗口作为输出流的任务。

### payload生成与迭代

```shell
-pl-gen   # payload 生成器
-pl-proc  # payload 处理器
-w        # fuzz 字典
-pl-dedup # 对生成的payload列表去重
-iter     # 指定使用的迭代器
```

和payload相关的设置总共有这些，`-w`参数的详细用法上面已经介绍过了，这里不赘述。

`-iter`参数用于指定使用的迭代器，内置的迭代器有4种：`clusterbomb`、`pitchfork`、`pitchfork-cycle`与`sniper`，它们的作用如下：

+ `clusterbomb`模式：枚举不同占位符对应的payload列表的所有组合。

+ `pitchfork`模式：对每个占位符payload列表使用相同的下标，遍历到最短的payload列表结束。

+ `pitchfork-cycle`模式：`pitchfork`模式的改进版本，每个占位符的列表下标仍同步更新，但是较短的列表结束后，下标会从0再循环，直到最长的列表结束。

+ `sniper`模式：**仅用于**单个占位符在请求中出现多次的情况。工具根据占位符出现的位置依次将特定位置的占位符替换为payload列表中的所有payload，并将其它位置的占位符替换为空。

除内置外，迭代器也可基于插件实现。

`-pl-dedup`对生成的payload列表进行去重。

`-pl-gen`参数用来指定fuzz占位符的payload生成器，可与`-w`参数共存。`-pl-gen`参数与`-w`参数类似，使用`::`符号关联payload生成器列表和占位符。payload生成器的命令行参数值遵循 [伪函数调用表达式](#插件调用) 语法。

目前工具内置3种payload生成器：

- `int(lower, upper, base, minLen)`: 生成 `[lower, upper)` 范围内指定 `base` 进制的数字字符串。e.g., `int(1, 100, 10)` 生成 "1" 到 "99"，`base`参数可省略，默认为10；`minLen`参数指定了生成的数字最小的位数，可省略，不足位数会使用前导0补足。
- `permute(s, maxLen)`: 生成字符串 `s` 的所有排列组合，最多 `maxLen` 个。e.g., `permute("abc", 10)`，若`maxLen`省略或为-1，则不限制。
- `permuteex(s, m, n)`: 生成字符串`s`的长度从`m`到`n`的全排列，若`n`未设置或小于0，则设置为最大长度。
- `nil(length)`: 生成一个长度为`length`，全为空字符串的列表。

`-pl-proc`参数用于指定fuzz占位符的对应payload使用的处理器，同样遵循 [伪函数调用表达式](#插件调用) 语法，使用`::`符号与占位符进行关联。
内置的6种payload处理器如下：

- `base64`: Base64 编码 payload。
- `urlencode`: URL 编码 payload。
- `addslashes`: 添加反斜杠转义特殊字符 (如 `'` -> `\'`)。
- `stripslashes`: 去除开头的 `/` 并将连续多个 `/` 替换为单个 `/`。
- `suffix(s)`: 给 payload 添加后缀 `s`。e.g., `suffix(".php")`。
- `repeat(n)`: 将 payload 重复 `n` 次。e.g., `repeat(3)` 将 "a" 变为 "aaa"。

**注意**：

+ 指定fuzz占位符时**不允许任何占位符是另一个占位符的子串**（比如指定了`FUZZ`，然后再指定`FUZZ1`），这种情况会导致模板解析失败，因此会拒绝执行
+ 若占位符没有绑定的payload生成器或字典，尝试对其绑定payload处理器会导致错误
+ [递归模式](#递归任务)与sniper模式下仅允许单个fuzz占位符存在

### http api配置

```shell
-api-addr     http api server listen address (default: 0.0.0.0:14514)
-api-tls      run http api server on https (default: false)
-http-api     enable http api mode (default: false)
-tls-cert-file        tls cert file to be used
-tls-cert-key tls cert key file to be used
```

工具允许以http服务的形式运行，在这种模式下，工具通过监听http请求来获取所要执行的任务。

若要启动http服务模式，可在命令行中指定`http-api`标志，使用`api-addr`选项指定http服务监听的地址。

`api-tls`标志用于启用https模式。`tls-cert-file`与`tls-cert-key`用于指定http模式使用的tls证书与tls密钥，若选项留空，采用`./fuzzgiu.cert`与`./fuzzgiu.key`作为默认值。

### 插件参数

```shell
  -preproc      preprocessor plugins to be used
  -preproc-prior-gen    preprocessor plugins to be usedbefore generating payloads
  -react        reactor plugin to be used
```

这类参数用于指定预处理器与反应器插件，具体用法可查看下文 [插件类型](#插件类型与作用简介) 的对应部分。

## tview界面操作

工具默认使用tview窗口作为任务输出窗口，在这个窗口中可以查看输出，或者对任务进行一些简单的操作比如暂停、恢复、退出等。

tview窗口分两类，一类是每个任务的信息窗口，包含任务进度、输出以及任务配置等信息

![单个任务](https://raw.githubusercontent.com/nostalgist134/FuzzGIU/refs/heads/main/imgs/eachJob.png)

任务信息窗口的控制方式如下：

+ 使用`Ctrl+W`/`Ctrl+K`/`Ctrl+Up`切换到当前选中的窗口的上窗口
+ 使用`Ctrl+S`/`Ctrl+J`/`Ctrl+Down`切换到当前选中的窗口的下窗口
+ 使用`方向键`/`hjkl`/`鼠标滚轮`在选中的窗口内滑动
+ 若选中的窗口为输出窗口或日志窗口，按下`c`键清空选中窗口的内容
+ 若选中的窗口为输出窗口或日志窗口，按下`Ctrl+L`键将窗口锁定到最新的输出，`Ctrl+U`恢复
+ 在任意窗口中按下`q`结束任务执行，按下`p`/`r`暂停或恢复任务的执行
+ 在任意窗口中按下`Ctrl+C`直接结束程序
+ 按下`Ctrl+R`切换到任务列表窗口中

另一类是任务列表窗口，列出了所有运行过/正在运行的任务以及进度信息

![任务列表](https://raw.githubusercontent.com/nostalgist134/FuzzGIU/refs/heads/main/imgs/listJobs.png)

在任何任务信息窗口中，均可按下`Ctrl+R`切换到任务列表窗口中；在任务列表中使用方向键上下可以选择任务，并回车切换到选中任务的任务信息窗口。

**注意**：

+ 在任务列表中，运行完毕的任务其尾部会加上`(done)`标签，不会自动从列表中移除，也可查看其任务信息。若有需要，可进入任务信息窗口后按下`q`键移除。同时，使用`q`键退出的任务，其任务列表项也会被移除。
+ 在http api模式启用的情况下，不能提交使用tview窗口输出的任务；单个任务不能同时使用tview与原生stdout作为输出流。
+ 若提交过使用tview窗口的任务，不能再提交包含原生stdout输出的任务，否则可能导致命令行窗口混乱。

## 进阶用法

### 递归任务

递归模式是FuzzGIU用于深度探测目标的高级功能，适用于需要逐层挖掘资源的场景（如目录枚举、多级路径探测等）。通过`-R`启用后，工具会根据响应结果动态生成新的 fuzz 任务，实现自动化的深度探测。

#### 核心参数详解

- `-R`：启用递归模式（仅支持单个 fuzz 占位符，避免多占位符导致的逻辑冲突）。
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

**注意**：

+ 单个任务的递归衍生任务在任务完全执行完毕后才会提交到任务池。
+ 单任务内部可以使用`delay`来控制发包速度，但目前版本下任务与任务之间没有牵制机制。若有需要，可使用`-c`控制最大并发任务数。

### 使用插件

FuzzGIU 通过插件系统扩展功能，支持自定义预处理、payload 生成、请求发送等逻辑，满足复杂场景的测试需求。插件为当前系统的共享库格式（`.so`/`.dll`/`.dylib`），需放置在`./plugins/[插件类型]/`目录下，若尚未创建插件目录，可不带任何参数运行一次工具，从而创建这些目录：

```powershell
PS H:\tools\fuzz\FuzzGIU> .\FuzzGIU.exe
checking/initializing environment...
Checking directory ./plugins/...Created.
Checking directory ./plugins/payloadGenerators/...Created.
Checking directory ./plugins/payloadProcessors/...Created.
Checking directory ./plugins/preprocessors/...Created.
Checking directory ./plugins/requesters/...Created.
Checking directory ./plugins/reactors/...Created.
Checking directory ./plugins/iterators/...Created.
done.
for help, use -h flag
```

#### 插件调用

FuzzGIU插件的命令行参数语法遵循以下的被称为**伪函数调用表达式**的规则：

1. 伪函数表达式的格式为：`fn_0([arg0, arg1, arg2, ...]),fn_1([arg0, arg1, ...]),...`。函数名`fn_n`即为使用的插件的名字，不同的函数间通过`,`隔开。若某个函数的参数列表为空，则括号也可省略
2. 表达式中的参数仅支持4种类型：`int`、`float64`、`bool`、`string`
3. 字符串参数使用单引号或者双引号括起来，两种都是可接受的，但是必须配对
4. `bool`型参数使用全小写的`true`和`false`
5. `int`型参数支持两种进制，10进制和16进制，默认按照10进制数算，但是如果无法解析为10进制（含有字母），则尝试解析为16进制数；也可显式指定16进制数，在数字前面加上`0x`前缀即可
6. 上下文参数：部分插件具有上下文参数，这些参数可能是fuzz过程中使用的**结构体**或**数据**，调用表达式中不指定，工具会根据任务上下文自动传递，不同插件的上下文参数参考[插件类型与作用简介](#插件类型与作用简介) 章节。模糊测试中使用的结构体的声明与作用可参考wiki中的[相关章节](https://github.com/nostalgist134/FuzzGIU/wiki/FuzzGIU部分实现细节#模糊测试相关结构体)。
7. 自定义参数：除上下文参数之外，插件开发者可以在插件中指定任意不违反操作系统限制的数量的自定义参数

无论是内置的组件还是自定义的插件都根据以上规则调用。函数名即为插件名，`./plugins/[插件类型]`为工具寻找插件动态链接库的目录。

#### 插件类型与作用简介

1. **Preprocessor（预处理插件）**

   + **作用**：在fuzz任务启动前对fuzz任务本身进行处理

   + **上下文参数**：`*fuzzTypes.Fuzz`->当前使用的fuzz任务
   + **返回值**：`*fuzzTypes.Fuzz`->处理后的fuzz任务

   + **链式调用**：若指定了多个预处理插件，在插件链上每一个插件的返回任务都会作为默认参数传递给下一个插件。

   示例：使用`job_dispatch`预处理插件将单个大型任务分布到多个主机上执行，并指定一个统一的结果收集服务器

   ```powershell
   .\FuzzGIU.exe -u http://test.com/FUZZ -w big_dict.txt::FUZZ -preproc job_dispatch("http://192.168.0.1:11451,http://192.168.0.2:11451","fyge4sYS+1I4TB54,zKx40QLm1t1gVMQf","http://192.168.0.3/results")
   ```
   
   **注意事项**：预处理插件前有两个调用点位，一个位于payload生成前，一个位于payload生成后，命令行参数中分别使用`-preproc-prior-gen`与`-preproc`指定。由于任务执行前会把payload列表以及总迭代长度也写入任务结构体中，因此`-preproc`也能感知到payload列表与迭代长度，`-preproc-prior-gen`则不行。

2. **PayloadGenerator（payload 生成器插件）**

   + **作用**：替代字典（`-w`）动态生成 payload，适用于规则化、定制化的测试场景（如 SQL 注入、XSS 测试用例）。
   + **上下文参数**：无
   + **返回值**：`[]string`->生成的payload
   + **链式调用**：指定多个payload生成器时，每个生成器生成的payload列表都会拼接到总列表中

   示例：`sqli`插件生成 SQL 注入测试 payload：

   ```powershell
   # 使用sqli插件生成注入payload，对ID参数进行fuzz
   .\FuzzGIU.exe -u http://test.com/?id=1FUZZ -pl-gen sqli::FUZZ
   ```

3. **PayloadProcessor（payload 处理器插件）**

   + **作用**：对生成的 payload 进行二次处理（如加密、编码），满足目标系统的格式要求。
   + **上下文参数**：`string`->要处理的payload，从占位符对应的payload列表中取出
   + **返回值**：`string`->处理后的payload
   + **链式调用**：若指定了多个payload处理器，则插件链上每个插件返回的处理后的payload都会作为默认参数传递给下一插件

   示例：`AES`插件对密码 payload 进行 AES 加密：

   ```powershell
   # 对密码字典中的payload先用AES加密，再base64编码
   .\FuzzGIU.exe -u http://test.com/login -d "user=admin&pass=PASS" -w pass.txt::PASS -pl-proc AES("1234567890abcdef"),base64::PASS
   ```

4. **Requester（请求发送插件）**

   + **作用**：扩展工具支持的协议
   + **上下文参数**：`*fuzzTypes.RequestCtx`->一个包含了请求本身和请求相关设置的上下文结构
   + **返回值**：`*fuzzTypes.Resp`->发送请求后接收到的响应结构
   + **链式调用**：不支持
   + **其它注意事项**：这类插件通过`-u`指定的url的scheme字段隐式调用，当`-u`指定的scheme不在工具预置支持的协议范围内，工具就会根据其scheme在`./plugins/requesters`目录中寻找对应名字的动态链接库。由于调用过程中不涉及伪函数表达式，因此此类插件无法接收自定义参数。

   示例：`ssh`插件用于 SSH 弱口令测试：

   ```powershell
   # 对SSH服务进行用户名/密码爆破（依赖ssh请求发送插件）
   .\FuzzGIU.exe -u ssh://USER:PASS@test.com:22 -w user.txt::USER -w pass.txt::PASS
   ```

5. **Reactor（响应处理器插件）**

   + **作用**：对请求和响应进行自定义分析（如指纹识别、漏洞特征匹配）并输出消息，或者根据结果来调控整个任务执行流（停止当前任务、生成新的请求/任务）。
   + **上下文参数**：`*fuzzTypes.Req`->请求结构体、`*fuzzTypes.Resp`->响应结构体
   + **返回值**：`*fuzzTypes.Reaction`->结构化的响应结果
   + **链式调用**：不支持

   示例：使用`fingerprint`插件识别目标服务器的指纹：

   ```powershell
   # 对响应进行指纹识别，输出服务器组件信息
   .\FuzzGIU.exe -u http://test.com/FUZZ -w endpoints.txt::FUZZ -react fingerprint
   ```

6. **Iterator（迭代器插件）**

   + **作用**：根据任务使用的占位符数量与每个占位符对应的payload列表决定迭代长度，或根据当前迭代下标选择要使用的占位符组合。
   + **链式调用**：本插件不支持链式调用

   迭代器由两个函数——`IterLen`与`IterIndex`共同组成。下面分别介绍每个函数的参数与返回值

   `IterLen`

   + **上下文参数**：`lengths []int`->每个占位符对应的payload列表的长度
   + **返回值**：正整数最大迭代长度，或者-1表示不限迭代次数

   `IterIndex`

   + **上下文参数**：`lengths []int`->每个占位符对应的payload列表长度、`ind int`->当前迭代下标
   + **返回值**：`indexes []int`->每个占位符使用的payload的下标

   **注意事项**：`IterIndex`返回的下标列表可以为负数，此时对应占位符会替换为空payload；当`IterLen`返回-1时，迭代不会自动停止，而是根据`IterIndex`返回值决定，当`IterInex`返回一个全负数的下标列表时停止。

#### 插件开发与扩展

若内置组件无法满足需求，可基于 [FuzzGIUPluginKit](https://github.com/nostalgist134/FuzzGIUPluginKit) 与go编译器（windows还需要gcc编译器）开发自定义插件，遵循工具定义的规范，实现特定逻辑后编译为动态链接库即可使用。

### http api端点

如上文所述，工具允许传入`-http-api`参数让自身以http服务模式运行。

api使用`Access-Token`头进行鉴权，`Access-Token`是长度为16的随机字符串，在工具以api模式启动时自动输出

```powershell
.\FuzzGIU.exe -http-api
listening at 0.0.0.0:14514
access token: mYOC23ROFTJVEAw+
```

访问所有的api端点都需要在请求头部带上`Access-Token: 工具输出的token`，否则访问失败，返回401状态码。

目前提供的api端点有4个，以下将分别介绍每个端点的作用。

#### job/:id

用法：根据任务id获取/停止某个正在运行的任务，允许`GET`/`DELETE`/`POST`方法，分别用于获取/停止任务。

根据方法不同，返回值如下：

+ `GET`：若任务id存在，则返回200状态码与json表示的任务信息，否则返回404状态码与错误信息。
+ `DELETE`：若任务id存在，则停止任务并返回204状态码，否则返回404状态码与错误信息。

##### `POST`方法

这个方法用于暂停/恢复任务的执行或者查看任务状态，调用时传入一个json格式的`map[string]string`，其`action`键值代表要执行的操作，可以为`pause`、`resume`或`status`，每种action的作用分别如下：

+ `pause`：暂停任务执行，返回204状态码
+ `resume`：恢复任务的执行，返回204状态码
+ `status`：返回任务状态，包括执行进度、执行过程中发生的错误总数以及任务是否暂停。

#### job

用法：用于提交json格式表示的`fuzzTypes.Fuzz`任务结构体，仅允许`POST`方法。

返回值：若任务提交成功，则返回200状态码与任务id，否则返回错误信息。

**注意**：

+ 提交的任务中如果使用插件，则需确保序列化为json时引用`fuzzTypes`包，因为`fuzzTypes.Plugin`类型采用了自定义的json序列化/反序列化函数，如果自行声明可能导致json反序列化失败。
+ http api模式下，不允许提交的任务使用`tview`窗口输出。

#### jobIds

用法：获取当前所有正在运行的任务的任务id列表，仅允许`GET`方法。

返回值：当前在运行任务的任务id列表。

#### stop

用法：以`GET`方法调用，停止Fuzzer的运行。

### libfgiu库

`libfgiu`（`github.com/nostalgist134/FuzzGIU/libfgiu`）库将`FuzzGIU`的核心功能封装成了对象与方法，允许在go代码中结合`fuzzTypes`等库调用`FuzzGIU`的功能进行模糊测试，并能够对测试任务进行一定程度上的控制与观测。

#### Fuzzer对象

`Fuzzer`对象用于执行与管理模糊测试任务，必须经由`libfgiu.NewFuzzer`方法分配后使用，典型的使用代码如下，更详细的使用方法可查看wiki

```go
package main

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/libfgiu"
	"log"
	"net/http"
	"time"
)

func main() {
	fuzzer, err := libfgiu.NewFuzzer(20, libfgiu.WebApiConfig{
		ServAddr:     "127.0.0.1:8080",
		TLS:          true,
		CertFileName: "certfile.cert",
		CertKeyName:  "certkey.key",
	}) // 获取一个fuzzer对象，并发数为20，并开启http api模式（若无需http api，省略第二个参数）
	if err != nil {
		log.Fatalln(err)
	}

	fuzzer.Start() // 启动fuzzer对象（启动fuzzer对象内部维护的并发任务池）
	fmt.Println("http api token:", fuzzer.GetApiToken())

	job := &fuzzTypes.Fuzz{ /*具体的任务配置*/ }
	jid, err := fuzzer.Submit(job) // 提交任务，并获取任务id
	if err != nil {
		log.Fatalln(err)
	}

	jobCtx, ok := fuzzer.GetJob(jid) // 根据任务id获取任务上下文并标记占用
	if ok {
		fmt.Printf("job#%d running\n", jid)
		jobInfo := jobCtx.GetJobInfo() // 即上面提交的任务（*fuzzTypes.Fuzz）
		fmt.Println("\tURL:", jobInfo.Preprocess.ReqTemplate.URL)

		jobCtx.Pause()   // 暂停任务执行
		// do something
		jobCtx.Resume()  // 继续任务的执行
		jobCtx.Release() // jobCtx内部维护一个引用计数，GetJob会增加此引用计数，手动释放避免任务在执行完成后阻塞
	}

	go func() {
		time.Sleep(100 * time.Second)
        fuzzer.StopHttpApi()
	}()

	fuzzer.Wait() // 等待直到fuzzer中全部任务都执行完毕
	fuzzer.Stop() // 停止fuzzer的运行
	if err != nil {
		log.Fatalln(err)
	}
	// 在通过/stop或者fuzzer.Stop停止fuzzer后，fuzzer对象不能再使用，必须调用NewFuzzer重新分配
}

```

# 特别感谢

## 项目

+ [ffuf/ffuf: Fast web fuzzer written in Go](https://github.com/ffuf/ffuf)
+ [yaklang/yakit: Cyber Security ALL-IN-ONE Platform](https://github.com/yaklang/yakit)
+ [Burp Suite - Application Security Testing Software - PortSwigger](https://portswigger.net/burp)

没有这些项目团队对于本类工具的探索与启发，就没有当前项目。

## 个人

特别感谢[@xch-space](https://github.com/xch-space)对项目命名提供的灵感。

