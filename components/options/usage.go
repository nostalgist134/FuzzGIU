package options

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func getSection(name string) string {
	switch {
	case name == "u" || name == "d" || name == "r" || name == "t" || name == "timeout" || name == "delay":
		return "general"
	case strings.HasPrefix(name, "m"):
		return "matcher"
	case strings.HasPrefix(name, "f"):
		return "filter"
	case name == "X" || name == "b" || name == "H" || name == "http2" || name == "F" || name == "s" || name == "x":
		return "HTTP"
	case strings.HasPrefix(name, "pl") || name == "w" || name == "mode":
		return "payload"
	case name == "o" || name == "fmt" || name == "v" || name == "ie" || name == "ns":
		return "output"
	case strings.HasPrefix(name, "rec") || name == "R":
		return "recursion"
	case strings.HasPrefix(name, "retry"):
		return "error handle"
	case name == "preproc" || name == "react":
		return "plugin"
	default:
		return "other"
	}
}

func exampleUsage(title string, cmdLines ...string) {
	fmt.Println(title + ":")
	for _, c := range cmdLines {
		fmt.Printf("    %s %s\n\n", os.Args[0], c)
	}
}

func usage() {
	fmt.Printf("Usage of %s:\n", os.Args[0])
	fmt.Printf("\t%s [options]\n", os.Args[0])
	fmt.Println("options are shown below. when fuzzGIU is executed without any args,\n" +
		"it will init and create plugin directory")
	grouped := map[string][]*flag.Flag{}

	// 遍历所有注册过的 flag
	flag.VisitAll(func(f *flag.Flag) {
		section := getSection(f.Name)
		grouped[section] = append(grouped[section], f)
	})
	// 分组打印
	for _, section := range []string{
		"general", "matcher", "filter", "HTTP", "payload", "output",
		"recursion", "error handle", "plugin", "other",
	} {
		flags := grouped[section]
		if len(flags) == 0 {
			continue
		}
		fmt.Fprintf(os.Stderr, "\n[%s settings]\n", section)
		for _, f := range flags {
			def := f.DefValue
			if def != "" {
				def = fmt.Sprintf(" (default: %s)", def)
			}
			fmt.Fprintf(os.Stderr, "  -%s\t%s%s\n", f.Name, f.Usage, def)
		}
	}
	fmt.Println("\nSIMPLE USAGES:")
	exampleUsage("fuzz URL", "-u http://test.com/FUZZ -w dict.txt::FUZZ",
		"-u http://test.com/MILAOGIU -w dict.txt  # use default keyword")
	exampleUsage("fuzz HTTP data", "-u http://test.com -w dict.txt::FUZZ -d \"test=FUZZ\"")
	exampleUsage("use embedded payload processor to process payload",
		"-u http://test.com -w dict.txt::FUZZ -d \"test=FUZZ\" "+
			"\\\n\t-pl-processor suffix(\".txt\"),base64::FUZZ  # base64 encode")
	exampleUsage("use embedded payload generators",
		"-u http://test.com/FUZZ \\\n\t"+
			"-pl-gen int(0,100,10)::FUZZ  # generate integer 0~100 with base 10")
	exampleUsage("use multiple fuzz keywords and keyword process mode",
		"-u http://FUZZ1/FUZZ2 -w dic1.txt::FUZZ1 \\\n\t-w dic2.txt::FUZZ2  # default mode is \"clusterbomb\"",
		"-u http://FUZZ3/FUZZ4 -w dic3.txt::FUZZ3 \\\n\t-w dic4.txt::FUZZ4 -m pitchfork-cycle")
	exampleUsage("use filters and matchers",
		"-w http://test.com/FUZZ -w dic.txt::FUZZ -mc 407 -fc 403-406 \\\n\t-ms 123-154 -fs 10-100,120")
	fmt.Println("refer to flag help information as above" +
		" or https://github.com/nostalgist134/FuzzGIU/wiki for more usages")
	fmt.Println("\nADVANCED USAGES:")
	exampleUsage("recursive jobs",
		"-u http://test.com/FUZZ -w dict.txt::FUZZ -R -rec-code 403 -rec-depth 4")
	exampleUsage("use plugins",
		"-u http://test.com/?id=FUZZ \\\n\t"+
			"-pl-gen sqli::FUZZ  # will search ./plugins/payloadGenerators/sqli.(so/dll/dylib)",
		"-u http://test.com -D \"name=admin&pass=PASS\" -w dict.txt::PASS "+
			"\\\n\t-pl-processor AES(\"1234567890abcdef\")::PASS  "+
			"# will search ./plugins/payloadProcessors/AES.(so/dll/dylib)",
		"-w user.txt::NAME -w pass.txt::PASS"+
			" \\\n\t-u ssh://USER:PASS@test.com:22  # ./plugins/requestSenders/ssh.(so/dll/dylib)",
		"-u http://test.com/FUZZ -w dict.txt::FUZZ "+
			"\\\n\t-preproc job_dispatch  # ./plugins/preprocessors/job_dispatch.(so/dll/dylib)",
		"-u http://test.com/FUZZ -w dict.txt::FUZZ "+
			"\\\n\t-react fingerprint  # plugins/reactors/fingerprint.(so/dll/dylib)")
	fmt.Printf(`
	when fuzzGIU is executed without any args, it will init and create plugin directory "./plugin" to refer to plugins. 
	there are 5 types of plugins can be used on current version: Preprocessor, PayloadGenerator, PayloadProcessor, 
	RequestSender and Reactor. every plugin is of shared library format of current operating system, fuzzGIU will try to 
	find plugin by name at ./plugin/pluginType, make sure you put the plugin file to the right directory. you can find 
	the usage of each type of plugin on https://github.com/nostalgist134/FuzzGIU/wiki. if you want to develop your own
	plugin, go check https://github.com/nostalgist134/FuzzGIUPluginKit, have fun :)`)
}
