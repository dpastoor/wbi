
# import urllib library
from urllib.request import urlopen

# import json
import json

def ConnectPackage(osstr):
    # store the URL in url as
    # parameter for urlopen
    url = "https://www.rstudio.com/wp-content/downloads.json"

    # store the response of URL
    response = urlopen(url)

    # storing the JSON response
    # from url in data
    data_json = json.loads(response.read())

    match osstr:
        case "U20":
            connect_os = "focal"
        case "U22":
            connect_os = "jammy"
        case "RH7":
            connect_os = "redhat7_64"
        case "RH8":
            connect_os = "redhat8"
        case "RH9":
            connect_os = "rhel9"
    match osstr:
        case "U20":
            pm_os = "focal"
        case "U22":
            pm_os = "jammy"
        case "RH7":
            pm_os = "redhat7_64"
        case "RH8":
            pm_os = "fedora28"
        case "RH9":
            pm_os = "rhel9"

    connect_url = data_json['connect']['installer'][connect_os]['url']
    pm_url = data_json['rspm']['installer'][pm_os]['url']

    # print the json response
    print(connect_url)
    print(pm_url)