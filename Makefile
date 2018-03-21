GO=go
default: ./
seeker:
	${GO} build  -tags nocgo -o seeker default
