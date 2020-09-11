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
	//"net/textproto"
	"net/http"
	"bufio"
	"strings"
	"io/ioutil"
	"compress/gzip"
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

	if oid, err := primitive.ObjectIDFromHex(hex.EncodeToString(key[:])); err == nil {
		return oid
	} else {
		log.WithError(err).Warn("failed to create object id")
		return primitive.NewObjectID()
	}
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

func DecodeHttpResponse(raw string) string {
	var header string
	trailer := "\n"
	reader := bufio.NewReader(strings.NewReader(raw))
	resp,err := http.ReadResponse(reader, &http.Request{})
	if err != nil{
		log.Info("Reading response: ",resp)
		return raw + trailer
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var bodyReader io.ReadCloser
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			bodyReader, err = gzip.NewReader(resp.Body)
			if err != nil {
				log.Error("Gunzipping body: ",err)
			}
			header  = "\n[==== GUNZIPPED ====]\n"
			trailer = "\n[===================]\n"
			defer bodyReader.Close()
		default:
			bodyReader = resp.Body
		}
		body, err := ioutil.ReadAll(bodyReader)
		if err != nil{
			log.Error("Reading body: ",err)
		}
		return raw + header + string(body) + trailer
	}
	return raw + trailer
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
