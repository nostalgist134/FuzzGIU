## 2026.1.4 v0.2.10
+ 优化了randMarker的计算
+ 修复了`-pl-gen`命令行参数无法正常解析的问题

## 2025.12.22 v0.2.9
+ 修复了协程池执行函数注册失败的问题

## 2025.12.14 v0.2.8
+ 重构了`fuzzTypes.PlGen`结构，现在单个关键字可同时使用payload生成器插件与字典作为生成源了
+ 移除`fuzzTypes.Resp`的`HttpResponse`字段，添加`StatCode`来表示状态码
+ 为`fuzzTypes`包中的对象添加了一些`receiver`

## 2025.12.12 v0.2.6
+ 添加了payload列表去重功能`-pl-dedup`

## 2025.12.11 v0.2.5
+ 修复了计数器显示错误数值bug
+ 为计数器添加了显示当前任务总输出数的字段
+ 将命令行参数`-v`的默认值从1改为2
+ 修复了http api返回无效的任务进度问题
+ 改进了`int`生成器，现在可以通过第四个参数来指定最小位数，不足的会使用前导0补足

## 2025.12.4 v0.2.4
+ 为`fuzzTypes.Match`添加了`MatchResponse`方法
+ 移除`stageReact`包的`matchResponse`函数，改为使用`Match.MatchResponse`
+ 移除`stageReact.React`函数中的一段无效if语句

## 2025.12.3 v0.2.3
+ 优化了tview窗口信息显示
+ 修复了iterator插件预加载错误的bug

## 2025.12.2 v0.2.2
+ 优化了tview窗口信息显示
+ 修复了tview窗口中按下`Ctrl+C`退出失败的问题
+ 修复了tview窗口的计数器在任务结束后有时无法显示正确任务数的问题
+ 将`github.com/nostalgist134/reusableBytes`依赖更新至`v0.1.3`版本
+ 更新了readme文件，解释tview窗口操作方法
+ 修复了递归功能无法正常使用的bug
+ 修复了`libfgiu.Fuzzer`对象漏掉衍生任务的bug
+ 优化了`libfgiu.Fuzzer.daemon`的逻辑，现在不会再出现假性任务提交了
+ 添加命令行`-c`选项，用于控制最大并发任务数
+ 移除了herobrine

## 2025.11.30 v0.2.0
+ 将fuzz包整体重构，现在支持多个fuzz任务并发
+ 暂时移除input功能以及对应函数
+ 添加新插件类型-`iterator`，具体使用方法查看readme与wiki
+ 为`fuzzTypes`包中的大部分对象添加了类方法，修改了一部分对象的结构与命名
+ 移除RunDirect、RunPassive，新增一个`libfgiu`包与可在其它go代码中使用的`libfgiu.Fuzzer`对象
+ 命令行输出窗口改用`tview`库实现，原先显示在输出窗口的logo移至帮助信息中
+ 添加了http api功能（即原先的被动模式）
+ 将`plugin.Plugin`移至`fuzzTypes`包中，自定义`Plugin`对象的反序列化与序列化，避免了any参数类型丢失问题
+ 修改`fuzzTypes`包中的声明，使其结构与命名更合理，并为大部分对象添加了receiver
+ 别的忘了

## 2025.09.16 v0.1.7
+ 添加了一个新的ReactFlag - `ReactMerge`，若在react插件返回值中指定了这一flag，则会将默认的react逻辑处理的结果与插件处理的结果进行归并（但以插件返回的reaction为主，也就是说不会覆盖插件已经填写的字段）
+ 将RunDirect、RunPassive作为导出函数，现在可通过调用这两个函数实现将FuzzGIU源码作为go库使用
+ 修复了http请求发送函数无法正确处理host头的bug
+ 添加了一个新的内置payload生成器`permuteex`，详细用法可见[readme](https://github.com/nostalgist134/FuzzGIU/blob/main/README.md#payload相关设置)

## 2025.09.09
更新，添加了一些新功能
+ 现在在ui界面中选中窗口后可以上下左右移动了，并且能将选中的窗口进行清空
+ 被动模式开放3个api，分别可用于添加任务、获取结果和获取当前任务；改善安全性，被动模式现在需要令牌访问
+ 现在windows下插件第一次调用失败时，会将缓冲区扩容至返回值的1.5倍
+ 建立`CHANGELOG.md`文件，之后更新内容不再在commit信息中介绍，而是同步到此文件

