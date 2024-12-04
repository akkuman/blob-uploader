package main

var (
    version = "dev"  // default "dev"，it will be overwrited by `-X main.version={{.Version}}`
    commit  = "none" // default "none"，it will be overwrited by `-X main.version={{.Commit}}`
    date    = "unknown" // 默认为 "unknown"，it will be overwrited by `-X main.version={{.Date}}`
)

func main() {
	Execute()
}