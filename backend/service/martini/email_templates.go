package martini

const approverEmailTemplate = `
<html xmlns="http://www.w3.org/1999/xhtml" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;">
  <head style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;">
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta http-equiv="X-UA-Compatible" content="IE=edge" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta name="format-detection" content="telephone=no" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta name="color-scheme" content="light only">
    <link href="http://fonts.cdnfonts.com/css/manrope" rel="stylesheet">
  </head>
  <style>
      :root {
          color-scheme: light;
      }
      body {
        height:100% !important;
        margin:0 !important;
        padding:0 !important;
        width:100% !important;
      }
      * {
        -webkit-font-smoothing: antialiased;
      }
  </style>
  <body style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; font-family: 'Manrope', Arial, Helvetica, sans-serif; margin: 0; min-width: 320px; text-size-adjust: none; color: #293033;">
    <table border="0" cellpadding="0" cellspacing="0" width="100%" style="border-collapse: collapse!important; background-color: #3FCAD4; background-image: linear-gradient(to top, #85D4C1, #3FCAD4);">
      <tbody>
        <tr>
          <td style="padding: 25px 0;">
            <table align="center" border="0" cellpadding="0" cellspacing="0" style="border-collapse: collapse!important;">
              <tbody>
                <tr>
                  <td>
                    <a href="https://www.martinisecurity.com" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; cursor: pointer; text-size-adjust: none;">
                        <img src="https://www.martinisecurity.com/static/mail_logo_light_r2.png" height="30" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; border: 0; box-sizing: border-box; display: block; text-size-adjust: none;"/>
                    </a>
                  </td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
        <tr>
          <td style="padding: 0 15px;">
            <table align="center" border="0" cellpadding="0" cellspacing="0" width="100%" style="border-collapse: collapse!important; max-width: 600px;">
              <tbody>
                  <tr>
                    <td style="padding: 40px 40px 50px 40px; -webkit-border-radius: 5px; -moz-border-radius: 5px; border-radius: 5px; background-color: #FFFFFF">
                      <table align="center" border="0" cellpadding="0" cellspacing="0" width="100%" style="border-collapse: collapse!important">
                        <tbody>
                          <tr>
                            <td style="text-align: left;">
                              <p style="font-size: 20px; font-style: normal; font-weight: bold; font-size: 20px; line-height: 30px; margin-bottom: 30px; letter-spacing: 0.2px">
                                Organization verification request
                              </p>
                              <p style="font-style: normal; font-weight: normal; font-size: 15px; line-height: 25px; letter-spacing: 0.4px; margin-bottom: 15px;">
                                Hello, {{.RequesterName}}! <br>{{.ApproverName}} ({{.ApproverEmail}}) has requested permission to acquire STIR/SHAKEN certificates for your organization:
                              </p>
                              <p style="font-style: normal; font-weight: 600; font-size: 15px; line-height: 25px; letter-spacing: 0.25px; margin-bottom: 30px;">
                                {{.Company}}
                                <br>
                                {{.Address}}
                              </p>
                              <p style="font-style: normal; font-weight: normal; font-size: 15px; line-height: 25px; letter-spacing: 0.4px;  margin-bottom: 15px;">
                                To approve this request you need to follow link below and enter the code that was provided you by the requestor.
                              </p>
                              <a href="https://{{.Hostname}}/validate/{{.Token}}" target="_blank" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box;cursor: pointer; text-decoration: none; text-size-adjust: none; word-break: break-all;">
                                <p style="font-style: normal; font-weight: normal; font-size: 15px; line-height: 25px; letter-spacing: 0.4px; text-decoration-line: underline; color: #377FF4; margin-bottom: 40px;">
                                    Approve request here
                                </p>
                              </a>
                              <p style="font-style: normal; font-weight: normal; font-size: 15px; line-height: 25px; letter-spacing: 0.4px; color: #8F999E;">
                                Cheers, <br>The team at Martini Security
                              </p>
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    </td>
                  </tr>
              </tbody>
            </table>
          </td>
        </tr>
        <tr style="-ms-text-size-adjust: none; -webkit-font-smoothing: subpixel-antialiased; -webkit-text-size-adjust: none;  box-sizing: border-box; text-size-adjust: none; text-align: center;">
          <td align="center" style="padding: 20px 15px 40px; font-size: 0px;">
            <table align="center" cellpadding="0" cellspacing="0" style="border-collapse:separate!important; border-spacing:0; box-sizing:border-box; padding:0; text-align:left; vertical-align:top; width:auto; max-width: 360px;">
              <tbody>
                <tr style="padding:0; vertical-align:top">
                  <td style="text-align:center">
                    <p style="color: #FFFFFF; -ms-text-size-adjust: none; -webkit-font-smoothing: subpixel-antialiased; -webkit-text-size-adjust: none; box-sizing: border-box; font-size: 11px; letter-spacing: 0.02em; line-height: 15px; margin: 0; text-align: center; text-size-adjust: none; word-break: break-word; max-width: 270px;">
                      This email contains sensitive information, please do not forward
                    </p>
                  </td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
      </tbody>
    </table>
  </body>
</html>
`

const requesterEmailTemplate = `
<html xmlns="http://www.w3.org/1999/xhtml" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;">
  <head style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;">
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta http-equiv="X-UA-Compatible" content="IE=edge" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta name="format-detection" content="telephone=no" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta name="color-scheme" content="light only">
    <link href="http://fonts.cdnfonts.com/css/manrope" rel="stylesheet">
  </head>
  <style>
      :root {
          color-scheme: light;
      }
      body {
        height:100% !important;
        margin:0 !important;
        padding:0 !important;
        width:100% !important;
      }
      * {
        -webkit-font-smoothing: antialiased;
      }
  </style>
  <body style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; font-family: 'Manrope', Arial, Helvetica, sans-serif; margin: 0; min-width: 320px; text-size-adjust: none; color: #293033;">
    <table border="0" cellpadding="0" cellspacing="0" width="100%" style="border-collapse: collapse!important; background-color: #3FCAD4; background-image: linear-gradient(to top, #85D4C1, #3FCAD4);">
      <tbody>
        <tr>
          <td style="padding: 25px 0;">
            <table align="center" border="0" cellpadding="0" cellspacing="0" style="border-collapse: collapse!important;">
              <tbody>
                <tr>
                  <td>
                    <a href="https://www.martinisecurity.com" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; cursor: pointer; text-size-adjust: none;">
                        <img src="https://www.martinisecurity.com/static/mail_logo_light_r2.png" height="30" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; border: 0; box-sizing: border-box; display: block; text-size-adjust: none;"/>
                    </a>
                  </td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
        <tr>
          <td style="padding: 0 15px;">
            <table align="center" border="0" cellpadding="0" cellspacing="0" width="100%" style="border-collapse: collapse!important; max-width: 600px;">
              <tbody>
                <tr>
                  <td style="padding: 40px 40px 50px 40px; -webkit-border-radius: 5px; -moz-border-radius: 5px; border-radius: 5px; background-color: #FFFFFF">
                    <table align="center" border="0" cellpadding="0" cellspacing="0" width="100%" style="border-collapse: collapse!important">
                      <tbody>
                        <tr>
                          <td style="text-align: left;">
                            <p style="font-size: 20px; font-style: normal; font-weight: bold; font-size: 20px; line-height: 30px; margin-bottom: 30px; letter-spacing: 0.2px;">
                              Organization verification submitted
                            </p>
                            <p style="font-style: normal; font-weight: normal; font-size: 15px; line-height: 25px; letter-spacing: 0.4px; margin-bottom: 25px;">
                              Hello, {{.RequesterName}}! <br>The organization verification request has been sent to {{.ApproverName}} ({{.ApproverEmail}}).
                            </p>
                            <p style="font-style: normal; font-weight: normal; font-size: 15px; line-height: 25px; letter-spacing: 0.4px; margin-bottom: 15px;">
                              Please provide the approver this code to complete the verification.
                            </p>
                            <table border="0" cellpadding="0" cellspacing="0" width="100%" style="border-collapse: collapse!important; margin-bottom: 30px;">
                              <tbody>
                                  <tr>
                                    <td style="padding: 13px 20px; background-color: #F4F7FC;">
                                      <p style="font-style: normal; font-weight: bold; font-size: 16px; line-height: 25px; letter-spacing: 0.15px;">
                                        {{.Code}}
                                      </p>
                                    </td>
                                  </tr>
                              </tbody>
                            </table>
                            <p style="font-style: normal; font-weight: normal; font-size: 15px; line-height: 25px; letter-spacing: 0.4px; color: #8F999E;">
                              Cheers, <br>The team at Martini Security
                            </p>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
        <tr style="-ms-text-size-adjust: none; -webkit-font-smoothing: subpixel-antialiased; -webkit-text-size-adjust: none;  box-sizing: border-box; text-size-adjust: none; text-align: center;">
          <td align="center" style="padding: 20px 15px 40px; font-size: 0px;">
            <table align="center" cellpadding="0" cellspacing="0" style="border-collapse:separate!important; border-spacing:0; box-sizing:border-box; padding:0; text-align:left; vertical-align:top; width:auto; max-width: 360px;">
              <tbody>
                <tr style="padding:0; vertical-align:top">
                  <td style="text-align:center">
                    <p style="color: #FFFFFF; -ms-text-size-adjust: none; -webkit-font-smoothing: subpixel-antialiased; -webkit-text-size-adjust: none; box-sizing: border-box; font-size: 11px; letter-spacing: 0.02em; line-height: 15px; margin: 0; text-align: center; text-size-adjust: none; word-break: break-word; max-width: 270px;">
                      This email contains sensitive information, please do not forward
                    </p>
                  </td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
      </tbody>
    </table>
  </body>
</html>
`

const orgApprovedTemplate = `

<html xmlns="http://www.w3.org/1999/xhtml" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;">
  <head style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;">
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta http-equiv="X-UA-Compatible" content="IE=edge" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta name="format-detection" content="telephone=no" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta name="color-scheme" content="light only">
    <link href="http://fonts.cdnfonts.com/css/manrope" rel="stylesheet">
  </head>
  <style>
      :root {
        color-scheme: light;
      }
      body {
        height:100% !important;
        margin:0 !important;
        padding:0 !important;
        width:100% !important;
      }
      * {
        -webkit-font-smoothing: antialiased;
      }
  </style>
  <body style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; font-family: 'Manrope', Arial, Helvetica, sans-serif; margin: 0; min-width: 320px; text-size-adjust: none; color: #293033;">
    <table border="0" cellpadding="0" cellspacing="0" width="100%" style="border-collapse: collapse!important; background-color: #3FCAD4; background-image: linear-gradient(to top, #85D4C1, #3FCAD4);">
      <tbody>
        <tr>
          <td style="padding: 25px 0;">
            <table align="center" border="0" cellpadding="0" cellspacing="0" style="border-collapse: collapse!important;">
              <tbody>
                <tr>
                  <td>
                    <a href="https://www.martinisecurity.com" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; cursor: pointer; text-size-adjust: none;">
                        <img src="https://www.martinisecurity.com/static/mail_logo_light_r2.png" height="30" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; border: 0; box-sizing: border-box; display: block; text-size-adjust: none;"/>
                    </a>
                  </td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
        <tr>
          <td style="padding: 0 15px;">
            <table align="center" border="0" cellpadding="0" cellspacing="0" width="100%" style="border-collapse: collapse!important; max-width: 600px;">
              <tbody>
                <tr>
                  <td style="padding: 40px 40px 50px 40px; -webkit-border-radius: 5px; -moz-border-radius: 5px; border-radius: 5px; background-color: #FFFFFF">
                    <table align="center" border="0" cellpadding="0" cellspacing="0" width="100%" style="border-collapse: collapse!important">
                      <tbody>
                        <tr>
                          <td style="text-align: left;">
                            <p style="font-size: 20px; font-style: normal; font-weight: bold; font-size: 20px; line-height: 30px; margin-bottom: 30px; letter-spacing: 0.2px;">
                              Organization verification succeeded!
                            </p>
                            <p style="font-style: normal; font-weight: normal; font-size: 15px; line-height: 25px; letter-spacing: 0.4px; margin-bottom: 40px;">
                              {{.Company}} is approved to acquire STIR/SHAKEN certificates. <br> Thank you for using Martini Security!
                            </p>
                            <p style="font-style: normal; font-weight: normal; font-size: 15px; line-height: 25px; letter-spacing: 0.4px; color: #8F999E;">
                              Cheers, <br>The team at Martini Security
                            </p>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
        <tr style="-ms-text-size-adjust: none; -webkit-font-smoothing: subpixel-antialiased; -webkit-text-size-adjust: none;  box-sizing: border-box; text-size-adjust: none; text-align: center;">
          <td align="center" style="padding: 20px 15px 40px; font-size: 0px;">
            <table align="center" cellpadding="0" cellspacing="0" style="border-collapse:separate!important; border-spacing:0; box-sizing:border-box; padding:0; text-align:left; vertical-align:top; width:auto; max-width: 360px;">
              <tbody>
                <tr style="padding:0; vertical-align:top">
                  <td style="text-align:center">
                    <p style="color: #FFFFFF; -ms-text-size-adjust: none; -webkit-font-smoothing: subpixel-antialiased; -webkit-text-size-adjust: none; box-sizing: border-box; font-size: 11px; letter-spacing: 0.02em; line-height: 15px; margin: 0; text-align: center; text-size-adjust: none; word-break: break-word; max-width: 270px;">
                      This email contains sensitive information, please do not forward
                    </p>
                  </td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
      </tbody>
    </table>
  </body>
</html>
`

const orgDeniedTemplate = `

<html xmlns="http://www.w3.org/1999/xhtml" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;">
  <head style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;">
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta http-equiv="X-UA-Compatible" content="IE=edge" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta name="format-detection" content="telephone=no" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; text-size-adjust: none;"/>
    <meta name="color-scheme" content="light only">
    <link href="http://fonts.cdnfonts.com/css/manrope" rel="stylesheet">
  </head>
  <style>
      :root {
        color-scheme: light;
      }
      body {
        height:100% !important;
        margin:0 !important;
        padding:0 !important;
        width:100% !important;
      }
      * {
        -webkit-font-smoothing: antialiased;
      }
  </style>
  <body style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; font-family: 'Manrope', Arial, Helvetica, sans-serif; margin: 0; min-width: 320px; text-size-adjust: none; color: #293033;">
    <table border="0" cellpadding="0" cellspacing="0" width="100%" style="border-collapse: collapse!important; background-color: #3FCAD4; background-image: linear-gradient(to top, #85D4C1, #3FCAD4);">
      <tbody>
        <tr>
          <td style="padding: 25px 0;">
            <table align="center" border="0" cellpadding="0" cellspacing="0" style="border-collapse: collapse!important;">
              <tbody>
                <tr>
                  <td>
                    <a href="https://www.martinisecurity.com" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; box-sizing: border-box; cursor: pointer; text-size-adjust: none;">
                        <img src="https://www.martinisecurity.com/static/mail_logo_light_r2.png" height="30" style="-ms-text-size-adjust: none; -webkit-text-size-adjust: none; border: 0; box-sizing: border-box; display: block; text-size-adjust: none;"/>
                    </a>
                  </td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
        <tr>
          <td style="padding: 0 15px;">
            <table align="center" border="0" cellpadding="0" cellspacing="0" width="100%" style="border-collapse: collapse!important; max-width: 600px;">
              <tbody>
                <tr>
                  <td style="padding: 40px 40px 50px 40px; -webkit-border-radius: 5px; -moz-border-radius: 5px; border-radius: 5px; background-color: #FFFFFF">
                    <table align="center" border="0" cellpadding="0" cellspacing="0" width="100%" style="border-collapse: collapse!important">
                      <tbody>
                        <tr>
                          <td style="text-align: left;">
                            <p style="font-size: 20px; font-style: normal; font-weight: bold; font-size: 20px; line-height: 30px; margin-bottom: 30px; letter-spacing: 0.2px;">
                              Organization verification denied!
                            </p>
                            <p style="font-style: normal; font-weight: normal; font-size: 15px; line-height: 25px; letter-spacing: 0.4px; margin-bottom: 40px;">
                              {{.Company}} is denied to acquire STIR/SHAKEN certificates. <br> Please contact approver for details.
                            </p>
                            <p style="font-style: normal; font-weight: normal; font-size: 15px; line-height: 25px; letter-spacing: 0.4px; color: #8F999E;">
                              Cheers, <br>The team at Martini Security
                            </p>
                          </td>
                        </tr>
                      </tbody>
                    </table>
                  </td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
        <tr style="-ms-text-size-adjust: none; -webkit-font-smoothing: subpixel-antialiased; -webkit-text-size-adjust: none;  box-sizing: border-box; text-size-adjust: none; text-align: center;">
          <td align="center" style="padding: 20px 15px 40px; font-size: 0px;">
            <table align="center" cellpadding="0" cellspacing="0" style="border-collapse:separate!important; border-spacing:0; box-sizing:border-box; padding:0; text-align:left; vertical-align:top; width:auto; max-width: 360px;">
              <tbody>
                <tr style="padding:0; vertical-align:top">
                  <td style="text-align:center">
                    <p style="color: #FFFFFF; -ms-text-size-adjust: none; -webkit-font-smoothing: subpixel-antialiased; -webkit-text-size-adjust: none; box-sizing: border-box; font-size: 11px; letter-spacing: 0.02em; line-height: 15px; margin: 0; text-align: center; text-size-adjust: none; word-break: break-word; max-width: 270px;">
                      This email contains sensitive information, please do not forward
                    </p>
                  </td>
                </tr>
              </tbody>
            </table>
          </td>
        </tr>
      </tbody>
    </table>
  </body>
</html>
`
