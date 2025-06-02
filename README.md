## Warning: This code is provided entirely for demonstration purposes and comes with zero warranty. Any use of this software is done so at the user's own risk

## Prerequisites
### Ngrok
For the asset processor tool to function correctly it is required that the service be externally accessible. To do this, ngrok is used to proxy incoming requests. For ngrok to allow for the service to be exposed on a consistent url, a paid account is required. If you have a paid account you will need an auth token and to add it as an "AUTH_TOKEN" environment variable in a `.env` file.

### Database Registration
To add your registration and deployment configuration you will need to insert the relevant rows into the postgres database.
The best way to do this is to create a file called `zdata.sql` inside the `db` folder. Inside the file, you will need to add a few SQL queries.
```
INSERT INTO key_set VALUES('d48a53de-021f-46f7-a0a4-7134812c2235');

INSERT INTO a_key VALUES(
    '1e3f0512-2066-4f8a-8916-2d278bf49524',
    'd48a53de-021f-46f7-a0a4-7134812c2235',
    '-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQD05ZgAT0Ysstqh
IJdzNcpl5N0R0jgDF6aY/4PZ7lWr/wHGxLvY/6Ta6QxEfAzHIGjrrqUwxmvdr8Rn
mG4FD0Wx7Qexe2HO9cLtVD+keLBOOz8C++dXvrugUZ9G4Ea2NE60nZgKDgPvNIJS
ki8Aq8oKHmvlZKFTCTQU+PlfMxa+tZ+9TcDfgjtpisBf00USXCuoSbAfqa+cHEYq
7lcU5nKo+HfO0kJTytGVjxTYvgw9C1hxnhMzof3+tmVlLVJyr15FTvxtBr8C7uMU
zbEP8fXCO+Eg2pL8vXH5834ia1OXG2SEUvEvxGwrASjkmU6/oJVXtvPcc6Y4cD/B
O98wXUuvAgMBAAECggEAVvi0xyNgJh6skz16W8GWLCKfuiUAnGSJV1ujDUHdrhQF
ovwaREHh35aVMzsk5JDngg/Hfa9x/kxeQXY6WFSoqTwnF8pcHX5dKCDb60KrRlU3
Fw20Bo7nnlNub+LVaf7glrdDGAsLLaflwoJE7AWWXoqYQeK/gjhKBUq4cS05Hie7
fg9lKOSFB7WXk76j/C9K2Ab7ah/6NbzmrV6mCruX1gPk19tG0Yw+10e5OIlLtxKT
55NDAiSq1/getURpE9SGi3ZPZcJzE0w28AjS2d8pfesRYROb5c8IaZygszPFEOcZ
EG57rXZtul5aCSUz87DXYcfJ8pJG4bzJdyWtYSL3TQKBgQD/eparwlaVZ9BRkZ5O
uD4Hho+QmcEdxeNrP0+UOW9wwsSV8rAZkQNsRg8Tx2JgcxO2Is+8y/5pmDpXJ+0W
S3xAcmIlbh4Xzvh7Eg7z+blB3SJnRqmleJSqLsG5Wlx4QptzH4UVUecmrnwIr7uE
XEpdj6rDi+L1MtF7ns8bgwtgbQKBgQD1ZXqwbKjTRXLhUGbNb+M1Eotr3JEKR+AN
6j9HRaH/NAe7sg9xNXHHjJ28Ihk/9cjmGXggCuL5N5cOflTnc5DTQr/uiebB8dlP
mGhlfinWvrCVryEpKPzuCq5eh5Y3CgidB6VuI72oYdGVAEsSWemr/0tl/f31Ksds
HUlkCS5jCwKBgHgpx76HzMO/LXuAO26ZOAvAHbyMpQmE7z+daqe0EBeAdIh2up97
1plRpnvOFxZ4afgMDZumc0ZlZGNkEx6eaJXDdyhVz++w2KzCRKg6eAljom/jC54Z
xgr5rQKqXr3tzkHqvGTXvhoyjYJkbZWG9y9kiJQrMpfTzDYR7yXokCxNAoGAXNmJ
w4lJk67aWdBPJXopUOJ0aFprcqVhbEJusOvy8JniNy2XVDFxnJxi7lVEkoPQAOgw
IIed+8gB2tUIEQ8UBCtkbcA11LpKjChRj91dvUgnjmtWM7mzgen+sfvBZY/hVHEZ
MgRJ9ZUVdLhIr2ff11lgUPX6ijImhIzMQRKMP6MCgYBw03imwmHKM0jIrHL+z9Gz
DfUc0xP30WfDF49WI2mEfotkuj4m8kZXOuz2FGMV/3WGB8lMhygfW29Uz3qs8YMj
/H2pj6pW16/IHFAiq7t+ZfJEsgjCyI7jOqyhCdlY1Ouue+bR4bl4oqIKhqZByrUV
YoDcMv8kd4KFhjvgdQPl8A==
-----END PRIVATE KEY-----',
    'RS256'
);

INSERT INTO registration VALUES(
    '56f3d0ed-0e0a-4ba5-a5a2-59aa4bbe6b57',
    'https://issuer.example.com',
    '457df601-695d-4ba6-8fbf-fef291ab3fb6',
    'https://example.com/oidc/login',
    'https://example.com/service/token',
    'https://example.com/.well-known/jwks',
    null,
    'https://lti-asset-processor-<user>.ngrok.io/lti/launch',
    'd48a53de-021f-46f7-a0a4-7134812c2235'
);
INSERT INTO deployment VALUES(
    '2394b381-8012-4310-98c4-76ef1d252157',
    '56f3d0ed-0e0a-4ba5-a5a2-59aa4bbe6b57',
    'Example Customer'
);
```