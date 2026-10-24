module gestionecanali

go 1.22

replace golang.org/x/sys => github.com/golang/sys v0.30.0

require (
	github.com/jchv/go-webview2 v0.0.0-20260205173254-56598839c808
	github.com/nwaples/rardecode/v2 v2.4.5
	golang.org/x/sys v0.0.0-20210218145245-beda7e5e158e
	golang.org/x/text v0.21.0
)

require github.com/jchv/go-winloader v0.0.0-20250406163304-c1995be93bd1 // indirect

replace golang.org/x/text => github.com/golang/text v0.21.0
