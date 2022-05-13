reindex:
	go run main.go reindexutxo

getbalance:
	go run main.go getbalance -address ${a}

send:
	go run main.go send -from ${f} -to ${t} -amount ${a}