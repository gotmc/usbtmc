package usbtmc

func invertbTag(bTag byte) byte {
	return bTag ^ 0xff
}
