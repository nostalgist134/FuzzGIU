package opt

import (
	"flag"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/version"
	"os"
	"strings"
)

func getSection(name string) string {
	switch {
	case name == "u" || name == "t" || name == "timeout" || name == "delay" ||
		name == "input" || name == "in-addr" || name == "passive" || name == "psv-addr" || name == "iter":
		return "GENERAL"
	case strings.HasPrefix(name, "m") && name != "mode":
		return "MATCHER"
	case strings.HasPrefix(name, "f") && name != "fmt":
		return "FILTER"
	case name == "X" || name == "b" || name == "H" || name == "http2" || name == "F" || name == "s" || name == "x" ||
		name == "ra" || name == "r" || name == "d":
		return "REQUEST"
	case strings.HasPrefix(name, "pl") || name == "w":
		return "PAYLOAD"
	case name == "out-file" || name == "fmt" || name == "v" || name == "ie" || name == "ns" || name == "tview" ||
		name == "out-url":
		return "OUTPUT"
	case strings.HasPrefix(name, "rec") || name == "R":
		return "RECURSION"
	case strings.HasPrefix(name, "retry"):
		return "RETRY"
	case name == "preproc" || name == "react":
		return "PLUGIN"
	case name == "http-api" || name == "api-tls" || name == "api-addr" || name == "tls-cert-file" ||
		name == "tls-cert-key":
		return "HTTP-API"
	default:
		return "OTHER"
	}
}

func exampleUsage(title string, execute bool, cmdLines ...string) {
	fmt.Println(title + ":")
	if !execute {
		for _, c := range cmdLines {
			fmt.Printf("    %s\n", c)
		}
		return
	}
	for _, c := range cmdLines {
		fmt.Printf("    %s %s\n", os.Args[0], c)
	}
	fmt.Println()
}

func usage() {
	fmt.Println(version.GetLogoVersionSlogan())
	fmt.Println("options are shown below. when FuzzGIU is executed without any args,\n" +
		"it will init and create plugin directory")
	grouped := map[string][]*flag.Flag{}

	// 遍历所有注册过的 flag
	flag.VisitAll(func(f *flag.Flag) {
		section := getSection(f.Name)
		grouped[section] = append(grouped[section], f)
	})
	// 分组打印
	for _, section := range []string{
		"GENERAL", "MATCHER", "FILTER", "REQUEST", "PAYLOAD", "OUTPUT",
		"RECURSION", "RETRY", "PLUGIN", "HTTP-API", "OTHER",
	} {
		flags := grouped[section]
		if len(flags) == 0 {
			continue
		}
		fmt.Fprintf(os.Stderr, "\n%s OPTIONS:\n", section)
		for _, f := range flags {
			def := f.DefValue
			if def != "" {
				def = fmt.Sprintf(" (default: %s)", def)
			}
			fmt.Fprintf(os.Stderr, "  -%s\t%s%s\n", f.Name, f.Usage, def)
		}
	}
	fmt.Println("\nSIMPLE USAGE EXAMPLES:")

	exampleUsage("fuzz URL", true, "-u http://test.com/FUZZ -w dict.txt::FUZZ",
		"-u http://test.com/MILAOGIU -w dict.txt  # use default keyword \"MILAOGIU\"")

	exampleUsage("fuzz request data", true,
		"-u http://test.com -w dict.txt::FUZZ -d \"test=FUZZ\"")

	exampleUsage("use filters and matchers", true,
		"-w http://test.com/FUZZ -w dic.txt::FUZZ -mc 407 -fc 403-406 \\\n\t-ms 123-154 -fs 10-100,120")

	exampleUsage("use embedded payload processor to process payload", true,
		"-u http://test.com -w dict.txt::FUZZ -d \"test=FUZZ\" "+
			"\\\n\t-pl-proc suffix(\".txt\"),base64::FUZZ  # first add .txt suffix, then base64 encode")

	exampleUsage("use embedded payload generators", true,
		"-u http://test.com/FUZZ \\\n\t"+
			"-pl-gen int(0,100,10)::FUZZ  # generate integer 0~100 with base 10",
		"-u http://test.com/FUZZ \\\n\t"+
			"-pl-gen permute('abcdefghijk',-1)::FUZZ # permutation of a-k",
		"-u http://test.com/FUZZ \\\n\t"+
			"-pl-gen permuteex('abcdefghijklm',2,3,-1)::FUZZ # permutation of a-m, length from 2 to 3")

	exampleUsage("use multiple fuzz keywords and iterator", true,
		"-u http://FUZZ1/FUZZ2 -w dic1.txt::FUZZ1 \\\n\t-w dic2.txt::FUZZ2  # default mode is \"clusterbomb\"",
		"-u http://FUZZ3/FUZZ4 -w dic3.txt::FUZZ3 \\\n\t-w dic4.txt::FUZZ4 -iter pitchfork-cycle")

	fmt.Println("\nADVANCED USAGE EXAMPLES:")

	exampleUsage("recursive jobs", true,
		"-u http://test.com/FUZZ -w dict.txt::FUZZ -R -rec-code 403 -rec-depth 4")

	exampleUsage("http api mode", false, "http api mode allow you to run FuzzGIU as an http "+
		"service to submit fuzz job via http request;\n    each fuzz job submitted are marked with an id, you can "+
		"use the URLs shown below to submit,\n    stop or inspect a job:\n"+
		"\tGET/DELETE /job/:id  -  get a JSON serialized job's structure by its id, or delete it\n"+
		"\tPOST /job  -  submit a job by serialized job structure, responses its id or error when failed\n"+
		"\tGET /jobIds  -  get all running job ids\n"+
		"\tGET /stop - stop the fuzzer\n")

	exampleUsage("use plugins", true,
		"-u http://test.com/?id=FUZZ \\\n\t"+
			"-pl-gen sqli::FUZZ  # will search ./plugins/payloadGenerators/sqli.(so/dll/dylib)",
		"-u http://test.com -d \"name=admin&pass=PASS\" -w dict.txt::PASS "+
			"\\\n\t-pl-proc AES(\"1234567890abcdef\")::PASS  "+
			"# will search ./plugins/payloadProcessors/AES.(so/dll/dylib)",
		"-w user.txt::USER -w pass.txt::PASS"+
			" \\\n\t-u ssh://USER:PASS@test.com:22  # ./plugins/requestSenders/ssh.(so/dll/dylib)",
		"-u http://test.com/FUZZ -w dict.txt::FUZZ "+
			"\\\n\t-preproc job_dispatch  # ./plugins/preprocessors/job_dispatch.(so/dll/dylib)",
		"-u http://test.com/FUZZ -w dict.txt::FUZZ "+
			"\\\n\t-react fingerprint  # ./plugins/reactors/fingerprint.(so/dll/dylib)",
		"-u http://test.com/FUZZ -w dict.txt::FUZZ "+
			"\\\n\t -iter random_index # ./plugins/iterators/random_index.(so/dll/dylib)")
	fmt.Println("FuzzGIU uses dynamic linked library as plugins to extend its functionality, " +
		"if you want to develop your own\nplugins, " +
		"go check https://github.com/nostalgist134/FuzzGIUPluginKit and wiki of FuzzGIU, have fun:)")
}
