#!/usr/bin/env python3

# fetch and generate go code for tracks information from https://github.com/bluemanos/forza-motorsport-car-track-ordinal

from urllib.request import urlopen
from csv import DictReader

URL = "https://raw.githubusercontent.com/bluemanos/forza-motorsport-car-track-ordinal/master/fm8/tracks.csv"
TRACKS = "generate/tracks.csv"
TRACKS_RESULT = "tracks_gen.go"

def fetch_tracks():
    with open(TRACKS, "wb") as f:
        with urlopen(URL) as response:
            if response.status != 200:
                print(response.read())
                raise RuntimeError("Request failed")
            f.write(response.read())

def generate_tracks():
    with open(TRACKS, "r") as f:
        tracks = list(DictReader(f, fieldnames=["ordinal","name","location","ioc_code","layout","length"]))

    tracks.sort(key=lambda t: int(t["ordinal"]))

    with open(TRACKS_RESULT, "w") as f:
        f.write(
    """// Code generated DO NOT EDIT.

package models

var Tracks = []Track{
"""
        )
        for track in tracks:
            f.write(
"""\t{{
\t\tOrdinal:  {ordinal},
\t\tName:     "{name}",
\t\tLayout:   "{layout}",
\t\tLocation: "{location}",
\t\tLength:   {length},
\t}},
""".format(**track))

        f.write("""}\n""")

if __name__ == "__main__":
    fetch_tracks()
    generate_tracks()
