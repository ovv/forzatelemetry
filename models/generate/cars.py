#!/usr/bin/env python3

# fetch and generate go code for tracks information from https://github.com/bluemanos/forza-motorsport-car-track-ordinal

from urllib.request import urlopen
from csv import DictReader

URL = "https://raw.githubusercontent.com/bluemanos/forza-motorsport-car-track-ordinal/master/fm8/cars.csv"
CARS = "generate/cars.csv"
CARS_RESULT = "cars_gen.go"

def fetch_cars():
    with open(CARS, "wb") as f:
        with urlopen(URL) as response:
            if response.status != 200:
                print(response.read())
                raise RuntimeError("Request failed")
            f.write(response.read())

def generate_cars():
    with open(CARS, "r") as f:
        cars = list(DictReader(f, fieldnames=["ordinal","year","make","model"]))

    cars.sort(key=lambda t: int(t["ordinal"]))

    with open(CARS_RESULT, "w") as f:
        f.write(
    """// Code generated DO NOT EDIT.

package models

var Cars = []Car{
"""
        )
        for car in cars:
            f.write(
"""\t{{
\t\tOrdinal: {ordinal},
\t\tYear:    {year},
\t\tMake:    "{make}",
\t\tModel:   "{model}",
\t}},
""".format(**car))

        f.write("""}\n""")

if __name__ == "__main__":
    fetch_cars()
    generate_cars()
