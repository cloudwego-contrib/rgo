install_lsp:
	go build -o go-lsp .
	mv go-lsp lsp/bin/go-lsp
