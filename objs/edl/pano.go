package edl

import (
    "fmt"
    "encoding/xml"

    "github.com/PaloAltoNetworks/pango/util"
    "github.com/PaloAltoNetworks/pango/version"
)

// PanoEdl is the panorama.Objects.Edl namespace.
type PanoEdl struct {
    con util.XapiClient
}

// Initialize is invoked when Initialize on the pango.Client is called.
func (c *PanoEdl) Initialize(con util.XapiClient) {
    c.con = con
}

// GetList performs GET to retrieve a list of EDLs.
func (c *PanoEdl) GetList(dg string) ([]string, error) {
    c.con.LogQuery("(get) list of EDLs")
    path := c.xpath(dg, nil)
    return c.con.EntryListUsing(c.con.Get, path[:len(path) - 1])
}

// ShowList performs SHOW to retrieve a list of EDLs.
func (c *PanoEdl) ShowList(dg string) ([]string, error) {
    c.con.LogQuery("(show) list of EDLs")
    path := c.xpath(dg, nil)
    return c.con.EntryListUsing(c.con.Show, path[:len(path) - 1])
}

// Get performs GET to retrieve information for the given EDL.
func (c *PanoEdl) Get(dg, name string) (Entry, error) {
    c.con.LogQuery("(get) EDL %q", name)
    return c.details(c.con.Get, dg, name)
}

// Get performs SHOW to retrieve information for the given EDL.
func (c *PanoEdl) Show(dg, name string) (Entry, error) {
    c.con.LogQuery("(show) EDL %q", name)
    return c.details(c.con.Show, dg, name)
}

// Set performs SET to create / update one or more EDLs.
func (c *PanoEdl) Set(dg string, e ...Entry) error {
    var err error

    if len(e) == 0 {
        return nil
    }

    _, fn := c.versioning()
    names := make([]string, len(e))

    // Build up the struct with the given configs.
    d := util.BulkElement{XMLName: xml.Name{Local: "external-list"}}
    for i := range e {
        d.Data = append(d.Data, fn(e[i]))
        names[i] = e[i].Name
    }
    c.con.LogAction("(set) EDLs: %v", names)

    // Set xpath.
    path := c.xpath(dg, names)
    if len(e) == 1 {
        path = path[:len(path) - 1]
    } else {
        path = path[:len(path) - 2]
    }

    // Create the objects.
    _, err = c.con.Set(path, d.Config(), nil, nil)
    return err
}

// Edit performs EDIT to create / update an EDL.
func (c *PanoEdl) Edit(dg string, e Entry) error {
    var err error

    _, fn := c.versioning()

    c.con.LogAction("(edit) EDL %q", e.Name)

    // Set xpath.
    path := c.xpath(dg, []string{e.Name})

    // Create the objects.
    _, err = c.con.Edit(path, fn(e), nil, nil)
    return err
}

// Delete removes the given EDLs from the firewall.
//
// EDLs can be either a string or an Entry object.
func (c *PanoEdl) Delete(dg string, e ...interface{}) error {
    var err error

    if len(e) == 0 {
        return nil
    }

    names := make([]string, len(e))
    for i := range e {
        switch v := e[i].(type) {
        case string:
            names[i] = v
        case Entry:
            names[i] = v.Name
        default:
            return fmt.Errorf("Unsupported type to delete: %s", v)
        }
    }
    c.con.LogAction("(delete) EDLs: %v", names)

    path := c.xpath(dg, names)
    _, err = c.con.Delete(path, nil, nil)
    return err
}

/** Internal functions for this namespace **/

func (c *PanoEdl) versioning() (normalizer, func(Entry) (interface{})) {
    v := c.con.Versioning()

    if v.Gte(version.Number{8, 0, 0, ""}) {
        return &container_v2{}, specify_v2
    } else {
        return &container_v1{}, specify_v1
    }
}

func (c *PanoEdl) details(fn util.Retriever, dg, name string) (Entry, error) {
    path := c.xpath(dg, []string{name})
    obj, _ := c.versioning()
    _, err := fn(path, nil, obj)
    if err != nil {
        return Entry{}, err
    }
    ans := obj.Normalize()

    return ans, nil
}

func (c *PanoEdl) xpath(dg string, vals []string) []string {
    if dg == "" {
        dg = "shared"
    }

    if dg == "shared" {
        return []string{
            "config",
            "shared",
            "external-list",
            util.AsEntryXpath(vals),
        }
    }

    return []string{
        "config",
        "devices",
        util.AsEntryXpath([]string{"localhost.localdomain"}),
        "device-group",
        util.AsEntryXpath([]string{dg}),
        "external-list",
        util.AsEntryXpath(vals),
    }
}