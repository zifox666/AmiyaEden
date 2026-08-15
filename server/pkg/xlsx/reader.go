package xlsx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

// Workbook is a deliberately small, dependency-free XLSX reader. It reads
// cached cell values only, which is exactly what published score workbooks
// must contain; formulas are never evaluated by the server.
type Workbook struct {
	Sheets map[string]map[string]string
}

func Open(data []byte) (*Workbook, error) {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	files := map[string]*zip.File{}
	if len(reader.File) > 500 {
		return nil, errors.New("XLSX 文件条目过多")
	}
	var uncompressed uint64
	for _, file := range reader.File {
		uncompressed += file.UncompressedSize64
		if uncompressed > 100*1024*1024 {
			return nil, errors.New("XLSX 解压内容过大")
		}
		files[path.Clean(file.Name)] = file
	}
	shared, err := readSharedStrings(files["xl/sharedStrings.xml"])
	if err != nil {
		return nil, err
	}
	workbookData, err := readZipFile(files["xl/workbook.xml"])
	if err != nil {
		return nil, errors.New("XLSX 缺少 xl/workbook.xml")
	}
	var workbookXML struct {
		Sheets []struct {
			Name  string `xml:"name,attr"`
			RelID string `xml:"id,attr"`
		} `xml:"sheets>sheet"`
	}
	if err := xml.Unmarshal(workbookData, &workbookXML); err != nil {
		return nil, err
	}
	relData, err := readZipFile(files["xl/_rels/workbook.xml.rels"])
	if err != nil {
		return nil, errors.New("XLSX 缺少工作簿关系文件")
	}
	var relXML struct {
		Relationships []struct {
			ID     string `xml:"Id,attr"`
			Target string `xml:"Target,attr"`
		} `xml:"Relationship"`
	}
	if err := xml.Unmarshal(relData, &relXML); err != nil {
		return nil, err
	}
	rels := map[string]string{}
	for _, rel := range relXML.Relationships {
		target := strings.TrimPrefix(rel.Target, "/")
		if !strings.HasPrefix(target, "xl/") {
			target = path.Join("xl", target)
		}
		rels[rel.ID] = path.Clean(target)
	}
	result := &Workbook{Sheets: map[string]map[string]string{}}
	for _, sheet := range workbookXML.Sheets {
		sheetFile := files[rels[sheet.RelID]]
		if sheetFile == nil {
			continue
		}
		cells, err := readSheet(sheetFile, shared)
		if err != nil {
			return nil, fmt.Errorf("读取工作表 %s: %w", sheet.Name, err)
		}
		result.Sheets[sheet.Name] = cells
	}
	return result, nil
}

func (w *Workbook) Cell(sheet, reference string) string {
	return strings.TrimSpace(w.Sheets[sheet][strings.ToUpper(reference)])
}

func readSharedStrings(file *zip.File) ([]string, error) {
	if file == nil {
		return nil, nil
	}
	data, err := readZipFile(file)
	if err != nil {
		return nil, err
	}
	var table struct {
		Items []struct {
			Text string `xml:"t"`
			Runs []struct {
				Text string `xml:"t"`
			} `xml:"r"`
		} `xml:"si"`
	}
	if err := xml.Unmarshal(data, &table); err != nil {
		return nil, err
	}
	result := make([]string, len(table.Items))
	for i, item := range table.Items {
		result[i] = item.Text
		for _, run := range item.Runs {
			result[i] += run.Text
		}
	}
	return result, nil
}

func readSheet(file *zip.File, shared []string) (map[string]string, error) {
	data, err := readZipFile(file)
	if err != nil {
		return nil, err
	}
	var sheet struct {
		Rows []struct {
			Cells []struct {
				Ref    string `xml:"r,attr"`
				Type   string `xml:"t,attr"`
				Value  string `xml:"v"`
				Inline string `xml:"is>t"`
			} `xml:"c"`
		} `xml:"sheetData>row"`
	}
	if err := xml.Unmarshal(data, &sheet); err != nil {
		return nil, err
	}
	result := map[string]string{}
	for _, row := range sheet.Rows {
		for _, cell := range row.Cells {
			value := cell.Value
			switch cell.Type {
			case "s":
				index, parseErr := strconv.Atoi(value)
				if parseErr != nil || index < 0 || index >= len(shared) {
					return nil, fmt.Errorf("无效共享字符串索引 %s", value)
				}
				value = shared[index]
			case "inlineStr":
				value = cell.Inline
			}
			result[strings.ToUpper(cell.Ref)] = value
		}
	}
	return result, nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	if file == nil {
		return nil, io.ErrUnexpectedEOF
	}
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}
