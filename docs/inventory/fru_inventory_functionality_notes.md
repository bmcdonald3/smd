# Inventory

## FRU Inventory

### How HSM Collects FRU Data

HSM's discovery mechanism walks specific sets of redfish URLs it finds for
components based of component HMSType and manufacturer. For nodeBMC HMSTypes,
HSM looks at these collections under /redfish/v1:
 - /redfish/v1/Chassis - Enclosure data
 - /redfish/v1/Systems - Node data
 - /redfish/v1/Managers - BMC data

From the Chassis collection generally just one entry is followed to collect FRU
data for the nodeEnclosure HMSType ('enclosure' if there are more than one in
the collection.)

Similarly, the Managers collection is used to collect BMC FRU data which
generally is the same as the node enclosure's FRU information. This is really
only collected to complete the hierarchy of component FRUs.

The Systems collection is used for collecting FRU info for nodes. Under the
/redfish/v1/Systems/<ID> URL are the processors (which contains processors and
GPUs), memory, and drives. FRU data is also collected by following those URLs.

Some node types also have node specific components that HSM discovers at the
Chassis level such as riser cards, HSN NICs, and GPUs (for proliant firmware).
These are generally found by following the links under the
/redfish/v1/Chassis/<systemID> URL:
 - /redfish/v1/Chassis/<systemID>/NetworkAdapters - HSN NICs
 - /redfish/v1/Chassis/<systemID>/Assembly - Riser cards
 - /redfish/v1/Chassis/<systemID>/Devices - GPUs (iLO proliant firmware only) 

HSM assigns each discovered FRU a FRUID based on the FRU data available.
Generally, HSM constructs the FRUID using:
'<HSMType>.<Manufacturer>.<PartNumber>.<SerialNumber>'

A FRU must have a serial number and a part number or manufacturer to be
trackable. Otherwise, HSM constructs a bogus FRUID for the FRU using
'FRUIDfor<componentID>'.

### When HSM Collects FRU Data

HSM is triggered to walk the redfish tree for a BMC under these conditions:
 - A 'POST /Inventory/Discover' is issued but only if 'Enabled' is true.
 - The RedfishEndpoint fields, 'RediscoverOnUpdate' and 'Enabled', are true and:
 -- A redfish endpoint is added to HSM via a 'POST /Inventory/RedfishEndpoints'.
 -- A redfish endpoint is modified in HSM via 'PATCH /Inventory/RedfishEndpoints'.
 - A partial discovery (/redfish/v1/Systems/<systemID> and below only) is performed when a node is indicated to have powered on via HSM receiving a redfish event.

### Database Storage

FRUs are stored in HSM's database in 2 parts, locational data and FRU data.
These are in the 'hwinv_by_loc' and 'hwinv_by_fru' tables respectively. The
entries in the 'hwinv_by_loc' table contain location specific hardware
non-specific data such as component ID, redfish ordinal, slot, etc. The entries
in the 'hwinv_by_fru' table contain the FRU info such as model, serial number,
part number manufacturer, and the FRUID HSM assigned to the FRU.

The 'hwinv_by_loc' table uses the FRUID as a foreign key so the entry in
'hwinv_by_fru' must exist. 'hwinv_by_loc' and 'hwinv_by_fru' can be detached if
the FRU is removed from the system or a discovered location is marked as absent
via redfish. 

## FRU Tracking

### FRU History Events

FRU history events are records of What/Where/When for FRU data. Current known
customer requirements for FRU tracking are just that a FRU must be trackable as
it gets moved to different locations in the system and that this data should
persist for up to a year.

### When Events Are Generated

Whenever FRU data is collected, HSM will attempt to generate a FRU history
event. HSM does this by matching the component ID and FRUID from the recently
collected FRU data with the component ID and FRUID of the previous FRU history
event for that component ID (location). If the FRUIDs don't match meaning the
FRU is new to the location, a FRU history event is generated. Otherwise, no
event is generated.

### Database Considerations

The FRU history events are stored in the 'hwinv_hist' table in HSM's database.
To keep the table a reasonable size, events that are stored are only for
indicating a change in the FRU's location. This is as of HSM version 1.30.7.

Previously, HSM generated events anytime FRU data was collected. This caused
the table size to grow to 20GB+ on 2k-4k node systems. Because of that,
migration step 20 runs a FRU history event pruning function to identify and
delete redundant FRU history events keeping only the changes.
