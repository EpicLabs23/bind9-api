package utils

import (
	"bind9-api/config"
	"bind9-api/models"
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/miekg/dns"
)

func UpdateZoneConfig(configFile string, zoneName string, zoneFile string) error {
	// Check if the zone already exists in the config
	file, err := os.OpenFile(configFile, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	zoneExists := false
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), `zone "`+zoneName+`"`) {
			zoneExists = true
			break
		}
	}

	if zoneExists {
		return nil
	}

	// Read old content to revert in case of syntax error
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	oldContent := make([]byte, fileInfo.Size())
	_, err = file.ReadAt(oldContent, 0)
	if err != nil {
		return err
	}

	// Append the new zone configuration
	zoneConfig := fmt.Sprintf(`
zone "%s" {
    type master;
    file "%s";
};`, zoneName, zoneFile)

	if _, err := file.WriteString(zoneConfig); err != nil {
		return err
	}

	// Syntax check the zone file
	success, output, _ := ExecCmd("named-checkconf", configFile)
	if !success {
		file.WriteString(string(oldContent)) // Revert to old content;
		return fmt.Errorf("config syntax error: %s", output)
	}

	return nil
}

func DeleteZoneFile(zoneFileDir string, zoneName string) error {
	filePath := filepath.Join(zoneFileDir, zoneName+".zone")
	return os.Remove(filePath)
}

func RemoveZoneFromConfig(configFile string, zoneName string) error {
	// This is a simplified implementation
	// In a real application, you'd want to properly parse the config file
	// and remove only the relevant zone block

	input, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(input), "\n")
	var newLines []string
	inZoneBlock := false

	for _, line := range lines {
		if strings.Contains(line, `zone "`+zoneName+`"`) {
			inZoneBlock = true
			continue
		}

		if inZoneBlock && strings.Contains(line, "};") {
			inZoneBlock = false
			continue
		}

		if !inZoneBlock {
			newLines = append(newLines, line)
		}
	}

	output := strings.Join(newLines, "\n")
	return os.WriteFile(configFile, []byte(output), 0644)
}

func ListZones(zoneFileDir string) ([]string, error) {
	var zones []string

	// Check if the directory exists
	if _, err := os.Stat(zoneFileDir); os.IsNotExist(err) {
		return zones, nil // Return empty slice if directory doesn't exist
	}

	files, err := os.ReadDir(zoneFileDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".zone") {
			zoneName := strings.TrimSuffix(file.Name(), ".zone")
			zones = append(zones, zoneName)
		}
	}

	return zones, nil
}

func ReloadBind9() (bool, string, error) {
	return ExecCmd("rndc", "reload")
}
func CheckCofigSyntax(configFile string) (bool, string, error) {
	success, output, err := ExecCmd("named-checkconf", configFile)
	if !success {
		return false, output, fmt.Errorf("config syntax error: %s", err.Error())
	}
	return success, output, err
}

func CheckAndReload(zoneName string, config *config.Config) (bool, string, error) {
	// Check zone syntax
	zoneFilePath := filepath.Join(config.Bind9.ZoneFileDir, zoneName+".zone")
	success, output, err := ExecCmd("named-checkzone", zoneName, zoneFilePath)
	if !success {
		return false, output, fmt.Errorf("zone syntax error: %s", err.Error())
	}

	// Check config syntax
	success, output, err = CheckCofigSyntax(config.Bind9.ConfigFile)
	if !success {
		return false, output, err
	}

	// Reload Bind9
	success, output, err = ExecCmd("rndc", "reload")
	if !success {
		return false, output, fmt.Errorf("failed to reload Bind9: %s", err.Error())
	}

	return success, output, err
}

func ParseZoneFile(filename string, origin string) ([]models.ResourceRecord, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open zone file: %w", err)
	}
	defer file.Close()

	var records []models.ResourceRecord
	reader := bufio.NewReader(file)

	zp := dns.NewZoneParser(reader, origin, filename)

	for rr, ok := zp.Next(); ok; rr, ok = zp.Next() {
		hdr := rr.Header()
		record := models.ResourceRecord{
			Name:   hdr.Name,
			TTL:    hdr.Ttl,
			Class:  dns.ClassToString[hdr.Class],
			Type:   dns.TypeToString[hdr.Rrtype],
			Fields: make(map[string]string),
		}

		switch v := rr.(type) {
		case *dns.A:
			record.Fields["address"] = v.A.String()
		case *dns.AAAA:
			record.Fields["address"] = v.AAAA.String()
		case *dns.CNAME:
			record.Fields["target"] = v.Target
		case *dns.NS:
			record.Fields["ns"] = v.Ns
		case *dns.MX:
			record.Fields["preference"] = fmt.Sprint(v.Preference)
			record.Fields["exchange"] = v.Mx
		case *dns.TXT:
			record.Fields["text"] = fmt.Sprint(v.Txt)
		case *dns.SOA:
			record.Fields["ns"] = v.Ns
			record.Fields["mbox"] = v.Mbox
			record.Fields["serial"] = fmt.Sprint(v.Serial)
			record.Fields["refresh"] = fmt.Sprint(v.Refresh)
			record.Fields["retry"] = fmt.Sprint(v.Retry)
			record.Fields["expire"] = fmt.Sprint(v.Expire)
			record.Fields["minttl"] = fmt.Sprint(v.Minttl)
		case *dns.SRV:
			record.Fields["priority"] = fmt.Sprint(v.Priority)
			record.Fields["weight"] = fmt.Sprint(v.Weight)
			record.Fields["port"] = fmt.Sprint(v.Port)
			record.Fields["target"] = v.Target
		case *dns.PTR:
			record.Fields["ptr"] = v.Ptr
		default:
			record.Fields["raw"] = v.String()
		}

		records = append(records, record)
	}

	if err := zp.Err(); err != nil {
		return nil, fmt.Errorf("zone parse error: %w", err)
	}

	return records, nil
}

func WriteZoneFile(zone models.Zone, conf *config.Config) error {
	zoneFileDir := conf.Bind9.ZoneFileDir
	origin := zone.Name
	ttl := zone.TTL
	if ttl == 0 {
		ttl = conf.Bind9.DefaultTTL
	}
	records := zone.Records

	// Ensure the zone directory exists
	if err := os.MkdirAll(conf.Bind9.ZoneFileDir, 0755); err != nil {
		return err
	}

	// Create the zone file
	filePath := filepath.Join(zoneFileDir, origin+".zone")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read old content to revert in case of syntax error
	oldContent, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Clear the file content
	if err := file.Truncate(0); err != nil {
		fmt.Print("Error truncating zone file:", err)
		return err
	}

	// Write $ORIGIN and $TTL
	if origin != "" {
		_, _ = fmt.Fprintf(file, "$ORIGIN %s.\n", strings.TrimSuffix(origin, "."))
	}
	if ttl > 0 {
		_, _ = fmt.Fprintf(file, "$TTL %d\n\n", ttl)
	}

	// Write each record
	for _, rr := range records {

		rrStr := MakeRRString(rr)
		record, err := dns.NewRR(rrStr)
		if err != nil {
			if len(oldContent) == 0 {
				os.Remove(filePath)
			} else {
				os.WriteFile(filePath, oldContent, 0644) // Revert to old content
			}
			return fmt.Errorf("zone syntax error: %s", err.Error())
		}

		// Handle @ shorthand for origin
		if rr.Name == origin || rr.Name == "" {
			rr.Name = origin
		}

		_, _ = fmt.Fprintf(file, "%s\n", record.String())
	}

	// Syntax check the zone file
	success, _, err := ExecCmd("named-checkzone", zone.Name, filePath)
	if !success {
		if len(oldContent) == 0 {
			os.Remove(filePath)
		} else {
			os.WriteFile(filePath, oldContent, 0644) // Revert to old content
		}
		return fmt.Errorf("zone syntax error: %s", err.Error())
	}

	return nil
}

func MakeRRString(record models.ResourceRecord) string {
	switch record.Type {
	case "A":
		return fmt.Sprintf("%s %d IN A %s", record.Name, record.TTL, record.Fields["address"])
	case "AAAA":
		return fmt.Sprintf("%s %d IN AAAA %s", record.Name, record.TTL, record.Fields["address"])
	case "CNAME":
		return fmt.Sprintf("%s %d IN CNAME %s", record.Name, record.TTL, record.Fields["target"])
	case "NS":
		return fmt.Sprintf("%s %d IN NS %s", record.Name, record.TTL, record.Fields["ns"])
	case "MX":
		return fmt.Sprintf("%s %d IN MX %s %s", record.Name, record.TTL, record.Fields["preference"], record.Fields["exchange"])
	case "TXT":
		return fmt.Sprintf("%s %d IN TXT \"%s\"", record.Name, record.TTL, record.Fields["text"])
	case "SOA":
		return fmt.Sprintf("%s %d IN SOA %s %s (%s %s %s %s %s)", record.Name, record.TTL, record.Fields["ns"], record.Fields["mbox"], record.Fields["serial"], record.Fields["refresh"], record.Fields["retry"], record.Fields["expire"], record.Fields["minttl"])
	case "SRV":
		return fmt.Sprintf("%s %d IN SRV %s %s %s %s", record.Name, record.TTL, record.Fields["priority"], record.Fields["weight"], record.Fields["port"], record.Fields["target"])
	case "PTR":
		return fmt.Sprintf("%s %d IN PTR %s", record.Name, record.TTL, record.Fields["ptr"])
	default:
		return fmt.Sprintf("%s %d IN %s %s", record.Name, record.TTL, record.Type, record.Fields["raw"])
	}
}
