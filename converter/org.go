package converter

type OrgFile struct {
    RealPath      string
    HTMLPath      string
    WebPath       string
    ID            string 
    Title         string
    LinkedToIDs   []string
}

var OrgFiles []OrgFile

func GetOrg(inputFile string) *OrgFile {
    for _, of := range OrgFiles {
        if of.RealPath == inputFile {
           return &of 
        }
    }
    return nil
}

func RmOrg(inputFile string) {
    var newOrgs []OrgFile
    for _, of := range OrgFiles {
        if of.RealPath != inputFile {
            newOrgs = append(newOrgs, of)
        }
    }
    OrgFiles = newOrgs
}

func(o *OrgFile) LinkedTo() []*OrgFile {
    var linkedTo []*OrgFile 
    for _, id := range o.LinkedToIDs {
        for i := range OrgFiles {
            if OrgFiles[i].ID == id {
                linkedTo = append(linkedTo, &OrgFiles[i])
            }
        }
    }
    return linkedTo
}

func(o *OrgFile) LinkedFrom() []*OrgFile {
    var linkedFrom []*OrgFile
    for i := range OrgFiles {
        for _, id := range OrgFiles[i].LinkedToIDs {
            if id == o.ID {
               linkedFrom = append(linkedFrom, &OrgFiles[i]) 
            }
        }  
    }
    return linkedFrom
}
