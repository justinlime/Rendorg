package converter

type OrgFile struct {
    RealPath      string
    HTMLPath      string
    WebPath       string
    ID            string 
    Title         string
    LinkedTo   []*OrgFile
    LinkedFrom []*OrgFile
}
