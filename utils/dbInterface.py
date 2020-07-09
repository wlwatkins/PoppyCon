import adafruit_ads1x15.ads1015 as ADS
from adafruit_ads1x15.analog_in import AnalogIn
from w1thermsensor import W1ThermSensor as w1
from peewee import *

from random import random
import time
import board
import busio

db = SqliteDatabase('data.db', pragmas={
    'journal_mode': 'wal',
    'cache_size': -1 * 64000,  # 64MB
    'foreign_keys': 1,
    'ignore_check_constraints': 0,
    'synchronous': 0})

class Sensors(Model):
    sensorType = CharField(max_length=200)
    sensorID = CharField(max_length=200)
    date = IntegerField()
    valueFloat = DecimalField(null=True)
    valueInt = IntegerField(null=True)
    name = CharField(max_length=100, null=True)
    desciption = CharField(max_length=500, null=True)

    class Meta:
        database = db

db.connect()
db.create_tables([Sensors])

def readHumidity():
    # Create the I2C bus
    i2c = busio.I2C(board.SCL, board.SDA)

    # Create the ADC object using the I2C bus
    adc1 = ADS.ADS1015(i2c, address=0x48)
    adc2 = ADS.ADS1015(i2c, address=0x49)

    data = {}

    probe = AnalogIn(adc1, ADS.P0)
    data["prob10"] =    {
                            "value": probe.value,
                            "voltage": probe.voltage,
                            "name": "prob10",
                            "desc": "prob10 description"
                        }

    probe = AnalogIn(adc1, ADS.P1)
    data["prob11"] =    {
                            "value": probe.value,
                            "voltage": probe.voltage,
                            "name": "prob11",
                            "desc": "prob11 description"
                        }

    probe = AnalogIn(adc1, ADS.P2)
    data["prob12"] =    {
                            "value": probe.value,
                            "voltage": probe.voltage,
                            "name": "prob12",
                            "desc": "prob12 description"
                        }

    probe = AnalogIn(adc1, ADS.P3)
    data["prob13"] =    {
                            "value": probe.value,
                            "voltage": probe.voltage,
                            "name": "prob13",
                            "desc": "prob13 description"
                        }

    probe = AnalogIn(adc2, ADS.P0)
    data["prob20"] =    {
                            "value": probe.value,
                            "voltage": probe.voltage,
                            "name": "prob20",
                            "desc": "prob20 description"
                        }

    probe = AnalogIn(adc2, ADS.P1)
    data["prob21"] =    {
                            "value": probe.value,
                            "voltage": probe.voltage,
                            "name": "prob21",
                            "desc": "prob21 description"
                        }

    probe = AnalogIn(adc2, ADS.P2)
    data["prob22"] =    {
                            "value": probe.value,
                            "voltage": probe.voltage,
                            "name": "prob22",
                            "desc": "prob22 description"
                        }

    probe = AnalogIn(adc2, ADS.P3)
    data["prob23"] =    {
                            "value": probe.value,
                            "voltage": probe.voltage,
                            "name": "prob23",
                            "desc": "prob23 description"
                        }
    return data

def readTemperature():
    data = {}
    i = 0
    for sensor in w1.get_available_sensors():
        data[sensor.id] = {"value": sensor.get_temperature(),
                            "name": f"temp {i}",
                            "desc": f"temp {i} desc",
                            }
        i+=1
    return data


def getMeasurements():
    data = {}
    data['moisture'] = readHumidity()
    data['temperature'] = readTemperature()

    return data
db.close()
if __name__ == "__main__":
    while True:
        db.connect()
        data = getMeasurements()
        now = int(time.time())
        for key, value in data['moisture'].items():
            record = Sensors.create(sensorType='MOISTURE',
                                    sensorID=key,
                                    date=now,
                                    valueFloat=value['voltage'],
                                    valueInt=int(value['value']),
                                    name=value['name'],
                                    desciption=value['desc'])

        for key, value in data['temperature'].items():
            record = Sensors.create(sensorType='TEMPERATURE',
                                    sensorID=key,
                                    date=now,
                                    valueFloat=value['value'],
                                    valueInt=-1,
                                    name=value['name'],
                                    desciption=value['desc'])
        db.close()
        time.sleep(60)
