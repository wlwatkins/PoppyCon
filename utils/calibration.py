import adafruit_ads1x15.ads1015 as ADS
from adafruit_ads1x15.analog_in import AnalogIn
from w1thermsensor import W1ThermSensor as w1
from utils.dbInterface import getMeasurements

from random import random
import time
import board
import busio


if __name__ == "__main__":
    while True:
        data = getMeasurements()
        now = int(time.time())
        print(data['moisture']["prob20"])
        # for key, value in data['moisture'].items():
        #     print(key, value)
