package golanglibs

import (
	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
	"github.com/alibabacloud-go/tea/tea"
)

type alidnsStruct struct {
	client          *alidns20150109.Client
	accessKeyID     string
	accessKeySecret string
}

type alidnsDomainInfoStruct struct {
	DomainName string
	DomainID   string
	CreateTime string
}

type alidnsDomainStruct struct {
	client     *alidns20150109.Client
	DomainName string
}

type alidnsRecord struct {
	ID   string
	Data string
	Name string
	TTL  int
	Type string
}

func getAlidns(accessKeyID string, accessKeySecret string) *alidnsStruct {
	config := &openapi.Config{
		AccessKeyId:     &accessKeyID,
		AccessKeySecret: &accessKeySecret,
	}

	config.Endpoint = tea.String("dns.aliyuncs.com")
	client, err := alidns20150109.NewClient(config)
	panicerr(err)

	return &alidnsStruct{
		client:          client,
		accessKeyID:     accessKeyID,
		accessKeySecret: accessKeySecret,
	}
}

func (m *alidnsStruct) total() (TotalCount int64) {
	describeDomainsRequest := &alidns20150109.DescribeDomainsRequest{}
	result, err := m.client.DescribeDomains(describeDomainsRequest)
	panicerr(err)
	TotalCount = *result.Body.TotalCount
	return
}

func (m *alidnsStruct) list(PageSize int64, PageNumber int64) (res []alidnsDomainInfoStruct) {
	describeDomainsRequest := &alidns20150109.DescribeDomainsRequest{
		PageSize:   &PageSize,
		PageNumber: &PageNumber,
	}
	result, err := m.client.DescribeDomains(describeDomainsRequest)
	panicerr(err)

	for _, d := range result.Body.Domains.Domain {
		res = append(res, alidnsDomainInfoStruct{
			DomainName: *d.DomainName,
			DomainID:   *d.DomainId,
			CreateTime: *d.CreateTime,
		})
	}
	return
}

func (m *alidnsStruct) domain(domainName string) *alidnsDomainStruct {
	return &alidnsDomainStruct{
		client:     m.client,
		DomainName: domainName,
	}
}

func (m *alidnsDomainStruct) list() (res []alidnsRecord) {
	result, err := m.client.DescribeDomainRecords(&alidns20150109.DescribeDomainRecordsRequest{
		DomainName: &m.DomainName,
	})
	panicerr(err)
	for _, r := range result.Body.DomainRecords.Record {
		res = append(res, alidnsRecord{
			ID:   *r.RecordId,
			Data: *r.Value,
			Name: *r.RR,
			TTL:  Int(*r.TTL),
			Type: *r.Type,
		})
	}
	return
}

func (m *alidnsDomainStruct) add(recordName string, recordType string, recordValue string) (id string) {
	recordType = String(recordType).Upper().Get()
	addDomainRecordRequest := &alidns20150109.AddDomainRecordRequest{
		DomainName: &m.DomainName,
		RR:         &recordName,
		Type:       &recordType,
		Value:      &recordValue,
	}
	res, err := m.client.AddDomainRecord(addDomainRecordRequest)
	panicerr(err)
	return *res.Body.RecordId
}

func (m *alidnsDomainStruct) delete(name string, dtype string, value string) {
	for _, d := range m.list() {
		var nameres, dtyperes, valueres bool = false, false, false
		if name == "" || d.Name == name {
			nameres = true
		}
		if dtype == "" || String(d.Type).Lower().Get() == String(dtype).Lower().Get() {
			dtyperes = true
		}
		if value == "" || d.Data == value {
			valueres = true
		}
		if nameres && dtyperes && valueres {
			_, err := m.client.DeleteDomainRecord(&alidns20150109.DeleteDomainRecordRequest{
				RecordId: &d.ID,
			})
			panicerr(err)
		}
	}
}

func (m *alidnsDomainStruct) modify(recordName string, srcRecordType string, srcRecordValue string, dstRecordName string, dstRecordType string, dstRecordValue string) {
	m.delete(recordName, srcRecordType, srcRecordValue)
	m.add(dstRecordName, dstRecordType, dstRecordValue)
}
