## 2025.09.16
+ 添加了一个新的ReactFlag - `ReactMerge`，若在react插件返回值中指定了这一flag，则会将默认的react逻辑处理的结果与插件处理的结果进行归并（但以插件返回的reaction为主，也就是说不会覆盖插件已经填写的字段�?
+ 将RunDirect、RunPassive作为导出函数，现在可通过调用这两个函数实现将FuzzGIU源码作为go库使�?
+ 修复了http请求发送函数无法正确处理host头的bug
+ 添加了一个新的内置payload生成器`permuteex`，详细用法可见[readme](https://github.com/nostalgist134/FuzzGIU/blob/main/README.md#payload相关设置)

## 2025.09.09
更新，添加了一些新功能�?
+ 现在在ui界面中选中窗口后可以上下左右移动了，并且能将选中的窗口进行清�?
+ 被动模式开�?个api，分别可用于添加任务、获取结果和获取当前任务；改善安全性，被动模式现在需要令牌访�?
+ 现在windows下插件第一次调用失败时，会将缓冲区扩容至返回值的1.5�?
+ 建立`CHANGELOG.md`文件，之后更新内容不再在commit信息中介绍，而是同步到此文件�?

