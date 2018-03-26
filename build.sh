# GO=go
# export GOPATH=$GOPATH:`pwd`
# seeker:
# 	${GO} build  -o seeker -tags nocgo ./
# export GOPATH=$GOPATH:`pwd`
# seeker:
# 	go build -o seeker ./
export GOPATH=$GOPATH:`pwd`
go build -o p2p-discovery2 ./
