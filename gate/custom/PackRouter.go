package customPack

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/gommon/log"
)

func ReadRouterConf(path string) error {
	fmt.Println(path)
	_processLine := func(line string) {
		sps := strings.Split(line, ":")
		if len(sps) != 3 {
			log.Error("ReadRouterConf len not 3")
			return
		}
		_id, _ := strconv.ParseUint(sps[0], 16, 16)
		id := uint16(_id)
		module := sps[1]
		_func := sps[2]
		router := PackRouter{
			Module: module,
			Id:     id,
			Func:   _func,
		}
		fmt.Println(router)
		RegisterPackRouter(id, &router)
	}
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		_processLine(line)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return nil
}
