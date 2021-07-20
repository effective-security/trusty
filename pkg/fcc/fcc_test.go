package fcc_test

import (
	"testing"

	"github.com/ekspand/trusty/pkg/fcc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var htmlTest string = `
<html>

<head>
	<title>FCC Registration System</title>
	<link rel="stylesheet" type="text/css" href="images/registration/cores.css">
</head>

<body>

<div id="srchDetailContent">


<div id="srchDetailClose"><a href="javascript:self.close()">Close Window</a></div>


<div id="srchDetailHeader">Registration Detail</div>       

<table>

<tr>
	<th>FRN:</th>
	<td>0000010827</td>
</tr>

<tr>
	<th>Registration Date:</th>
	<td>09/15/2000 08:30:58 AM</td>
</tr>

<tr>
	<th>Last Updated:</th>
	<td>
	          	
		04/05/2021 12:12:43 PM
	</td>
</tr>

<tr>
	<th>Business Name:</th>
	<td>	
	
		Veracity Networks, LLC
	
	</td>
</tr>

<!-- <##dba##> -->

<tr>
	<th>Business Type:</th>
	<td>
	          	
		Private Sector
										
	
	,
										
	          	
		Limited Liability Corporation
	
	</td>
</tr>

<tr>
	<th>Contact Organization:</th>
	<td>
	          	
		
										
	</td>
</tr>

<tr>
	<th>Contact Position:</th>
	<td>
	          	
		FCC Contact
										
	</td>
</tr>
<tr>
	<th>Contact Name:</th>
	<td>

	
	    Tara Lyle 
	

	</td>
</tr>

<tr>
	<th>Contact Address:</th>
	<td>
	  
	
	
357 S. 670 W.<br>
       

Ste 300<br>
       

Lindon, UT 84042<br>
       
       

United States                                      
                                            

	
	</td>
</tr>

<tr>
	<th>Contact Email:</th>
	<td>
	          	
		tara.lyle@veracitynetworks.com
	
	</td>
</tr>

<tr>
	<th>ContactPhone:</th>
	<td >
	          	
		(801) 878-3225 
	
	</td>
</tr>

<tr>
	<th>ContactFax:</th>
	<td>
	          	
		(801) 373-0682
										
	</td>
</tr>

</table>

</div> <!-- Close srchDetailContent -->

</body>
</html>
`

func TestParseEmailFromHTML(t *testing.T) {
	email, err := fcc.ParseEmailFromHTML([]byte(htmlTest))
	require.NoError(t, err)
	assert.Equal(t, "tara.lyle@veracitynetworks.com", email)
}

var xmlForTest string = `
<?xml version="1.0" encoding="ISO-8859-1"?>
<Filer499QueryResults xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:noNamespaceSchemaLocation="http://apps.fcc.gov/cgb/form499/XMLSchema/Filer499QueryResults_Schema.xsd"
 Updated="2021-06-21" RecordCount="1" >
		
        <Filer>
            <Form_499_ID>831188</Form_499_ID>
            <Filer_ID_Info>
                <Registration_Current_as_of>2021-04-01</Registration_Current_as_of>
                <start_date>2015-01-12</start_date>
                <USF_Contributor>Yes</USF_Contributor>
                <Legal_Name>LOW LATENCY COMMUNICATIONS LLC</Legal_Name>
                <Principal_Communications_Type>Interconnected VoIP</Principal_Communications_Type>
                <holding_company>IPIFONY SYSTEMS INC.</holding_company>
                <FRN>0024926677</FRN>
                <hq_address>
                    <address_line>241 APPLEGATE TRACE</address_line>
                    <city>PELHAM</city>
                    <state>AL</state>
                    <zip_code>35124</zip_code>
                </hq_address>
                <customer_inquiries_address>
                    <address_line>241 APPLEGATE TRACE</address_line>
                    <city>PELHAM</city>
                    <state>AL</state>
                    <zip_code>35124</zip_code>
                </customer_inquiries_address>
                <Customer_Inquiries_telephone>2057453970</Customer_Inquiries_telephone>
                <other_trade_name>Low Latency Communications</other_trade_name>
                <other_trade_name>String by Low Latency</other_trade_name>
                <other_trade_name>Lilac by Low Latency</other_trade_name>
            </Filer_ID_Info>
            <Agent_for_Service_Of_Process>
                <dc_agent>Jonathan Allen Rini O&apos;Neil, PC</dc_agent>
                <dc_agent_telephone>2029553933</dc_agent_telephone>
                <dc_agent_fax>2022962014</dc_agent_fax>
                <dc_agent_email>jallen@rinioneil.com</dc_agent_email>
                <dc_agent_address>
                    <address_line>1200 New Hampshire Ave, NW</address_line>
                    <address_line>Suite 600</address_line>
                    <city>Washington</city>
                    <state>DC</state>
                    <zip_code>20036</zip_code>
                </dc_agent_address>
            </Agent_for_Service_Of_Process>
            <FCC_Registration_information>
                <Chief_Executive_Officer>Daryl Russo</Chief_Executive_Officer>
                <Chairman_or_Senior_Officer>Matthew Hardeman</Chairman_or_Senior_Officer>
                <President_or_Senior_Officer>Larry Smith</President_or_Senior_Officer>
            </FCC_Registration_information>
            <jurisdiction_state>alabama</jurisdiction_state>
            <jurisdiction_state>florida</jurisdiction_state>
            <jurisdiction_state>georgia</jurisdiction_state>
            <jurisdiction_state>illinois</jurisdiction_state>
            <jurisdiction_state>louisiana</jurisdiction_state>
            <jurisdiction_state>north_carolina</jurisdiction_state>
            <jurisdiction_state>pennsylvania</jurisdiction_state>
            <jurisdiction_state>tennessee</jurisdiction_state>
            <jurisdiction_state>texas</jurisdiction_state>
            <jurisdiction_state>virginia</jurisdiction_state>
        </Filer>
</Filer499QueryResults>
`

func TestParseFRNFromXML(t *testing.T) {
	frn, err := fcc.ParseFRNFromXML([]byte(xmlForTest))
	require.NoError(t, err)
	assert.Equal(t, "0024926677", frn)
}
