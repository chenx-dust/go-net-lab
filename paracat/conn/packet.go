package conn

import (
	"errors"
	"io"
)

const MAGIC_NUMBER = 0xa1

func WritePacket(writer io.Writer, buffer []byte, id uint16) (n int, err error) {
	packed := make([]byte, 0, 5+len(buffer))
	packed = append(packed, MAGIC_NUMBER)
	packed = append(packed, byte(len(buffer)))
	packed = append(packed, byte(len(buffer)>>8))
	packed = append(packed, byte(id))
	packed = append(packed, byte(id>>8))
	packed = append(packed, buffer...)

	n = 0
	for n < len(packed) {
		n_, err := writer.Write(packed[n:])
		if err != nil {
			return n + n_, err
		}
		n += n_
	}
	return
}

func ReadPacket(reader io.Reader, buffer []byte) (length int, id uint16, err error) {
	header := make([]byte, 5)
	n, err := reader.Read(header)
	if err != nil {
		return
	}
	if n < 5 {
		return 0, 0, errors.New("invalid packet")
	}

	if header[0] != MAGIC_NUMBER {
		return 0, 0, errors.New("invalid magic number")
	}

	length = int(header[1]) | int(header[2])<<8
	id = uint16(header[3]) | uint16(header[4])<<8
	pt := 0
	for pt < length {
		n, err = reader.Read((buffer)[pt:length])
		if err != nil {
			return
		}
		pt += n
	}
	return
}
