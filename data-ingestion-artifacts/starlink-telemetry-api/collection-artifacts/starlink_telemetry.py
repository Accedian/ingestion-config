import requests
import os
from pprint import pprint
from urllib.parse import urlencode
import json 
import time
import sys

last_query_timestamp = 0
query_interval_s = 30

def get_starlink_access_token():
  payload =  {
        'client_id' : "!!!!! REPLACE WITH YOUR CLIENT ID !!!!!!",
        'client_secret' : "!!!!! REPLACE WITH YOUR CLIENT SECRET !!!!!!",
        'grant_type' : 'client_credentials'
    }

  response = requests.post(
            'https://api.starlink.com/auth/connect/token',
            data= urlencode(payload),
            headers=
                {
                    'Content-Type' : 'application/x-www-form-urlencoded'
                }
        )
  
  if (response.status_code != 200):
    sys.exit("Auth failed")

  return response.json()['access_token']


def get_column_index(column_names, desired_column_name):
    """
    Gets the index for a telemtry metric (ex: ObstructionPercentTime).

    :param column_names: column names for a specific device type.
    :param desired_column_name: column you want to get index of (ex: ObstructionPercentTime).
    :return: index of the metric, or -1 if it can not be found.
    """
    index = 0
    for column_name in column_names:
        if column_name == desired_column_name:
            return index
        index += 1

    return -1

def poll_stream():
    """
    Constantly polls telemetry API. Expect to get a response about every 15 seconds.
    When called initially, you might receive a response more often as the stream catches up.
    """
    access_token = get_starlink_access_token()

    while (True):
        try:

            query_start = time.time()

            response = requests.post(
                'https://web-api.starlink.com/telemetry/stream/v1/telemetry',
                json=
                    {
                        "accountNumber": "ACC-4241022-76323-1",
                        "batchSize": 5000,
                        "maxLingerMs": 20000
                    },
                headers=
                    {
                        'content-type' : 'application/json',
                        'accept' : '*/*',
                        'Authorization' : 'Bearer '+access_token
                    }
            )

            if (response.status_code != 200):
                # Auth token expires ~15 minutes, so refresh it if invalid response.
                access_token = get_starlink_access_token()
            else:
                response_json = response.json()

                # The raw telemetry data points for all device types.
                telemetry = response_json['data']['values']

                # If no telemetry received, don't do any processing.
                if (len(telemetry) == 0):
                    continue

                # User terminal column names to figure out the index of telemetry in the raw data.
                ut_column_names = response_json['data']['columnNamesByDeviceType']['u']

                # Router column names to figure out the index of telemetry in the raw data.
                # router_column_names = response_json['data']['columnNamesByDeviceType']['r']

                # Mapping to human readable alert names for user terminals.
                # ut_alert_names = response_json['metadata']['enums']['AlertsByDeviceType']['u']
    
                json_data = [dict(zip(ut_column_names, data)) for data in telemetry]
                print(json.dumps(json_data,separators=(',', ':')))
        
                elapsed_time = time.time() - query_start
                if (elapsed_time > query_interval_s):
                    print("WARN: query time is greater than interval. will run again immediately. consider increasing interval time")
                sleep_time = max(0, query_interval_s - elapsed_time)
                time.sleep(sleep_time)
        except KeyboardInterrupt:
                sys.exit("Received interrupt. Exiting")

if __name__ == '__main__':
    poll_stream()