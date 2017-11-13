# The challenge

Goal​: Create a simple web application to which you can import drivers and add a rest API to
view the imported drivers.

Input: JSON file containing list of drivers:

```
[{"id":3,"name":"Johnny B. Goode","license_number":"12-234-45"},
{"id":1352,"name":"Taylor Swift","license_number":"44-455-10"},
{"id":2,"name":"Eyal Golan","license_number":"11-288-10"},
{"id":7,"name":"Janice Joplin","license_number":"65-112-10"},
{"id":9,"name":"Keren Peles","license_number":"12-234-45"},
{"id":8,"name":"Brittney Spears","license_number":"62-932-14"},
{"id":5,"name":"Michael Jackson","license_number":"12-224-42"},
{"id":4,"name":"Freddie Mercury","license_number":"46-251-01"}]
```

* First task​: Build a data model for the driver using PostgreSQL.
* Second task​: Build a simple service that provides the following API:
  * (POST) /import : Imports the drivers’ data into the datastore using the data model
implemented before.
  * (GET) /driver/<id> : Gets a driver's info (in a json format) by his id.

Use one of the supported languages (Golang, Python, Node.js, Java or C#).

You should build a service that is up and running (locally) and responds properly to calls via
browser, Postman or Curl.

Bonus Task​: Add API testing (in any language you prefer).

Bonus Task​: Upload the web app you developed to a cloud provider, such as
Heroku/DigitalOcean/AWS/google appengine.

Pay attention to:
- Correctness
- Code Quality over speed of implementation
- Project Structure / code organization
- Error handling

The code should be simple and meet the goal. No need for an "Enterprise Solution" here :-)