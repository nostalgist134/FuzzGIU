package version

import "fmt"

const logo = "     GIUGIUGIUGIU                            GIUGIUGIUGI GIUGIUGI     GI\n      GI                                    IU            GI  UG     UG\n     UG                                    GI            UG  IU     IU\n    IUGIUGIUGGI   IU#GIUGIUGIU#GIUGIUGIU# UG            IU  GI     GI\n   GI       UG   UG       GIU       GIU  IU       GIU  GI  UG     UG\n  UG       IU   GI      GIU       GIU   GI        IU  UG  IU     IU\n IU       GI   IU     GIU       GIU    UG        UG  IU  GI     GI\nGIU        UGIUGIU GIUGIUGIU#GIUGIUGIU#IUGIUGIUGIU GIUGIU UGIUGIU  %s   -   %s\n"

var (
	version = "v0.2.6"
	slogan  = "i have no ideas about slogans yet ):"
)

func GetVersion() string {
	return version
}

func GetLogoVersionSlogan() string {
	return fmt.Sprintf(logo, version, slogan)
}
