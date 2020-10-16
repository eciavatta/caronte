/*
 * This file is part of caronte (https://github.com/eciavatta/caronte).
 * Copyright (c) 2020 Emiliano Ciavatta.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, version 3.
 *
 * This program is distributed in the hope that it will be useful, but
 * WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
 * General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"net"
	"os"
	"time"
)

func Sha256Sum(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer func() {
		err = f.Close()
		if err != nil {
			log.WithError(err).WithField("filename", fileName).Error("failed to close file")
		}
	}()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func CustomRowID(payload uint64, timestamp time.Time) RowID {
	var key [12]byte
	binary.BigEndian.PutUint32(key[0:4], uint32(timestamp.Unix()))
	binary.BigEndian.PutUint64(key[4:12], payload)

	oid, err := primitive.ObjectIDFromHex(hex.EncodeToString(key[:]))
	if err == nil {
		return oid
	}
	log.WithError(err).Warn("failed to create object id")
	return primitive.NewObjectID()
}

func NewRowID() RowID {
	return primitive.NewObjectID()
}

func EmptyRowID() RowID {
	return [12]byte{}
}

func RowIDFromHex(hex string) (RowID, error) {
	rowID, err := primitive.ObjectIDFromHex(hex)
	return rowID, err
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func FileSize(filename string) int64 {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return -1
	}
	return info.Size()
}

func byID(id RowID) OrderedDocument {
	return OrderedDocument{{"_id", id}}
}

func DecodeBytes(buffer []byte, format string) string {
	switch format {
	case "hex":
		return hex.EncodeToString(buffer)
	case "hexdump":
		return hex.Dump(buffer)
	case "base32":
		return base32.StdEncoding.EncodeToString(buffer)
	case "base64":
		return base64.StdEncoding.EncodeToString(buffer)
	case "ascii":
		str := fmt.Sprintf("%+q", buffer)
		return str[1 : len(str)-1]
	case "binary":
		str := fmt.Sprintf("%b", buffer)
		return str[1 : len(str)-1]
	case "decimal":
		str := fmt.Sprintf("%d", buffer)
		return str[1 : len(str)-1]
	case "octal":
		str := fmt.Sprintf("%o", buffer)
		return str[1 : len(str)-1]
	default:
		return string(buffer)
	}
}

func CopyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	if err := in.Close(); err != nil {
		return err
	}
	return out.Close()
}

func ParseIPNet(address string) *net.IPNet {
	_, network, err := net.ParseCIDR(address)
	if err != nil {
		ip := net.ParseIP(address)
		if ip == nil {
			return nil
		}

		size := 0
		if ip.To4() != nil {
			size = 32
		} else {
			size = 128
		}
		network = &net.IPNet{
			IP:   ip,
			Mask: net.CIDRMask(size, size),
		}
	}

	return network
}

func Average(array []float64) float64 {
	var sum float64
	for _, f := range array {
		sum += f
	}
	return sum / float64(len(array))
}
