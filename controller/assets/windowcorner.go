package assets

import (
	"bytes"
	"image"
	"image/png"
)

func WindowCorner() image.Image {
	img, err := png.Decode(bytes.NewBuffer(pngFileData))
	if err != nil {
		panic(err)
	}
	return img
}

var pngFileData = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x1a, 0x08, 0x06, 0x00, 0x00, 0x00, 0xbe,
	0x68, 0xdc, 0x07, 0x00, 0x00, 0x00, 0x01, 0x73, 0x52, 0x47, 0x42, 0x00, 0xae, 0xce, 0x1c, 0xe9, 0x00, 0x00, 0x00, 0x04, 0x67, 0x41, 0x4d, 0x41, 0x00, 0x00, 0xb1, 0x8f, 0x0b, 0xfc,
	0x61, 0x05, 0x00, 0x00, 0x00, 0x09, 0x70, 0x48, 0x59, 0x73, 0x00, 0x00, 0x0e, 0xc3, 0x00, 0x00, 0x0e, 0xc3, 0x01, 0xc7, 0x6f, 0xa8, 0x64, 0x00, 0x00, 0x01, 0x4d, 0x49, 0x44, 0x41,
	0x54, 0x38, 0x4f, 0x95, 0x94, 0x3d, 0x4b, 0x03, 0x41, 0x10, 0x86, 0xf7, 0x27, 0x1c, 0x68, 0x30, 0x55, 0xe0, 0xfe, 0x87, 0x55, 0x04, 0x03, 0x06, 0x05, 0x53, 0x69, 0x0a, 0x9b, 0xc3,
	0x42, 0x0e, 0x8b, 0x10, 0x52, 0xa5, 0xb5, 0x4c, 0x99, 0xff, 0xbb, 0xee, 0x3b, 0x7b, 0xef, 0x3a, 0x3b, 0x99, 0x3b, 0x62, 0xf1, 0x70, 0x3b, 0x73, 0x33, 0xcf, 0xce, 0x7d, 0xb0, 0xa1,
	0x99, 0xb5, 0xf1, 0x26, 0x31, 0x9f, 0xb7, 0x71, 0x76, 0x97, 0xd7, 0xcd, 0x6d, 0x06, 0x6b, 0xe4, 0x98, 0x27, 0xf9, 0xfe, 0x42, 0x08, 0x2c, 0x84, 0x00, 0xfc, 0x15, 0xe4, 0xb5, 0x6e,
	0xe6, 0xfa, 0xf9, 0xa5, 0x8b, 0xfb, 0xc3, 0x59, 0xa8, 0x04, 0x28, 0x60, 0xf3, 0x18, 0xdd, 0xd7, 0x4f, 0xec, 0x53, 0x23, 0x58, 0x25, 0x51, 0xd0, 0x76, 0xee, 0xe0, 0x35, 0x82, 0xe5,
	0xe3, 0x9b, 0x08, 0xd0, 0x48, 0x02, 0x1b, 0xaf, 0x11, 0xa0, 0x19, 0xe8, 0x9c, 0x08, 0x38, 0x05, 0xd1, 0x05, 0x9a, 0xcd, 0xfb, 0x5e, 0xd0, 0xb9, 0xf2, 0x08, 0xc0, 0xca, 0x74, 0xac,
	0x05, 0xba, 0xa6, 0x12, 0x68, 0xbc, 0xc9, 0x5c, 0x81, 0x57, 0x68, 0x41, 0x8d, 0x37, 0x01, 0xf2, 0x17, 0x13, 0xa0, 0x90, 0x6b, 0xca, 0x79, 0xb5, 0x02, 0x7c, 0x7a, 0xf9, 0x0f, 0x3c,
	0xd8, 0xc8, 0x66, 0x2b, 0xe0, 0xbd, 0x49, 0x81, 0x95, 0xfc, 0x4b, 0x00, 0xc6, 0x04, 0xcc, 0x8d, 0x0a, 0x50, 0x4c, 0x3c, 0x81, 0xbe, 0xef, 0x7e, 0x05, 0x9b, 0xb3, 0x82, 0x26, 0xc5,
	0xdc, 0xac, 0x12, 0x60, 0xcd, 0xd8, 0x5e, 0x51, 0x5c, 0x04, 0x43, 0x8c, 0xeb, 0xe8, 0x7f, 0xa0, 0x65, 0xa4, 0x08, 0x86, 0xe6, 0x22, 0xb8, 0x56, 0x52, 0x4d, 0x30, 0x20, 0x3f, 0x92,
	0x3e, 0x8d, 0x3c, 0x58, 0x5c, 0x4d, 0x30, 0x50, 0x1e, 0xc1, 0xee, 0xa6, 0x99, 0x14, 0xb0, 0x88, 0x02, 0x2d, 0xb4, 0x82, 0x7e, 0x77, 0x8a, 0xab, 0xf5, 0x87, 0x2f, 0x00, 0xfa, 0xac,
	0x03, 0xdd, 0xf7, 0x49, 0xe8, 0x77, 0xe9, 0x08, 0x13, 0x26, 0x04, 0xf6, 0xac, 0x03, 0x4f, 0xaf, 0x99, 0xed, 0xe7, 0x51, 0x60, 0xac, 0x37, 0x15, 0x81, 0x77, 0xd6, 0x69, 0xc1, 0xfd,
	0xc3, 0x46, 0x70, 0x05, 0x78, 0x56, 0x9e, 0x75, 0x48, 0xe8, 0xf7, 0x00, 0x38, 0x2a, 0x63, 0xc2, 0x9a, 0x80, 0x4f, 0xc8, 0xb7, 0x6b, 0x1b, 0xb9, 0x66, 0x7c, 0x29, 0x6a, 0xe3, 0x2f,
	0xef, 0xa5, 0x5c, 0x49, 0x16, 0x01, 0x15, 0x32, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}