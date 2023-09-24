package agent

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/indece-official/monitor-agent-linux/src/generated/model/apiagent"
	"github.com/indece-official/monitor-agent-linux/src/utils"
	"gopkg.in/guregu/null.v4"
)

const CheckerTypeFile = "com.indece.agent.linux.v1.checker.file"

type FileChecker struct {
}

func (c *FileChecker) GetType() string {
	return CheckerTypeFile
}

func (c *FileChecker) md5sum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("can't open file for md5 sum: %s", err)
	}
	defer file.Close()

	md5sum := md5.New()
	_, err = io.Copy(md5sum, file)
	if err != nil {
		return "", fmt.Errorf("can't read file for md5 sum: %s", err)
	}

	return fmt.Sprintf("%x", md5sum.Sum(nil)), nil
}

func (c *FileChecker) GetChecker() (*apiagent.CheckerV1, error) {
	return &apiagent.CheckerV1{
		Name:    "File",
		Type:    CheckerTypeFile,
		Version: "",
		Params: []*apiagent.CheckerV1Param{
			{
				Name:     "path",
				Label:    "File path",
				Type:     apiagent.CheckerV1ParamType_CheckerV1ParamTypeText,
				Required: true,
			},
			{
				Name:  "calc_md5",
				Label: "Calc md5 sum of file",
				Type:  apiagent.CheckerV1ParamType_CheckerV1ParamTypeBoolean,
			},
		},
		Values: []*apiagent.CheckerV1Value{
			{
				Name:    "exists",
				Type:    apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
				MinCrit: "0",
			},
			{
				Name: "size",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeNumber,
			},
			{
				Name: "age",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeDuration,
			},
			{
				Name: "created_at",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeDateTime,
			},
			{
				Name: "md5",
				Type: apiagent.CheckerV1ValueType_CheckerV1ValueTypeText,
			},
		},
		CustomChecks: true,
		// Run only every 5 min
		DefaultSchedule: "0 */5 * * * *",
	}, nil
}

func (c *FileChecker) GetChecks() ([]*apiagent.CheckV1, error) {
	return []*apiagent.CheckV1{}, nil
}

func (c *FileChecker) Check(ctx context.Context, params []*apiagent.CheckV1Param) (string, []*apiagent.CheckV1Value, error) {
	paramPath := null.String{}
	paramCalcMD5 := false

	var err error

	for _, param := range params {
		if param.Value == "" {
			continue
		}

		switch param.Name {
		case "path":
			paramPath.Scan(param.Value)
		case "calc_md5":
			paramCalcMD5, err = strconv.ParseBool(param.Value)
			if err != nil {
				return "", nil, fmt.Errorf("error parsing parameter '%s': %s", param.Name, err)
			}
		default:
			return "", nil, fmt.Errorf("unknown parameter '%s'", param.Name)
		}
	}

	if !paramPath.Valid || paramPath.String == "" {
		return "", nil, fmt.Errorf("missing parameter 'path'")
	}

	values := []*apiagent.CheckV1Value{}

	exists := false
	size := int64(0)
	createdAt := null.Time{}
	messageError := ""

	fileStat, err := os.Stat(paramPath.String)
	if err != nil {
		messageError = err.Error()
	} else {
		exists = true
		size = fileStat.Size()
		createdAt.Scan(fileStat.ModTime())
	}

	values = append(values, &apiagent.CheckV1Value{
		Name:  "exists",
		Value: fmt.Sprintf("%d", utils.BoolToInt(exists)),
	})

	values = append(values, &apiagent.CheckV1Value{
		Name:  "size",
		Value: fmt.Sprintf("%d", size),
	})

	if createdAt.Valid {
		values = append(values, &apiagent.CheckV1Value{
			Name:  "created_at",
			Value: createdAt.Time.Format(time.RFC3339Nano),
		})

		values = append(values, &apiagent.CheckV1Value{
			Name:  "age",
			Value: time.Since(createdAt.Time).String(),
		})
	}

	md5sum := ""

	if exists && paramCalcMD5 {
		md5sum, err = c.md5sum(paramPath.String)
		if err != nil {
			messageError = err.Error()
		}
	}

	values = append(values, &apiagent.CheckV1Value{
		Name:  "md5",
		Value: md5sum,
	})

	message := ""

	if messageError == "" {
		message = fmt.Sprintf(
			"File %s found (%s)",
			paramPath.String,
			utils.FormatBytes(size),
		)
	} else {
		message = fmt.Sprintf(
			"Error checking file %s: %s",
			paramPath.String,
			messageError,
		)
	}

	return message, values, nil
}

var _ IChecker = (*FileChecker)(nil)

func NewFileChecker() *FileChecker {
	return &FileChecker{}
}
