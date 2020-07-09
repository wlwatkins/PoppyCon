import adafruit_ads1x15.ads1015 as ADS
from adafruit_ads1x15.analog_in import AnalogIn
from w1thermsensor import W1ThermSensor as w1
try:
    from utils.dbInterface import getMeasurements
except Exception as e:
    from dbInterface import getMeasurements
from random import random
import time
import board
import busio
import sys



if __name__ == "__main__":
    SENSORTYPE = {  "0": "moisture",
                    "1": "temperature"}
    if sys.argv[1] == "-h" or sys.argv[1] == "help":
        print("""
        python calibration.py X Y
            X: is the type of sensor
                0 => moisture
                1 => temperature
                2 => light
            Y: name of sensor
        To get list of sensors type "-L"
        """)
    elif sys.argv[1] == "-L":
        data = getMeasurements()
        for k, v in data.items():
            print("X:", k, f"(code arg: {[q for q, w in SENSORTYPE.items() if w == k]})")
            for k, v in data[k].items():
                print("     Y:", k)
    else:
        try:
            while True:
                data = getMeasurements()
                now = int(time.time())
                print(data[SENSORTYPE[sys.argv[1]]][sys.argv[2]])
        except Exception as e:
            print("Error accured", e)
