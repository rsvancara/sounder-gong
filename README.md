# Sounder Gong 

Has the  Covid19 pandemic blues and endless days of zoom meetings result in post traumatic stress 
disorder?  Well look no further! Liven up your zoom meetings with Sounder Gong, a small little application
that will add flare to otherwise boring zoom meetings through a convenient, easy to use interface which allows you to
upload your sound clips on a DAC powered raspberry pi.  The sound clips are accessible through a simple to use web API
that can be integrated with a variety of controllers that support calling a simple URL, such as the Elgato Stream Deck.

The reason I built this project includes bringing a better experience to zoom meetings that I host daily as part of my 
job.  Remote work and zoom has basically turned us all into media presenters, whether we like or not. 

## Design

Using a rasberry pi and a HiFiBerry DAC, the goal is to build a low latency, line out sound producing device that can 
be managed through a web interface with a web API with configurable links that can be configured as a button control
in an Elgato Stream Deck.  The Stream Deck already has integrations for calling URLs that can be configured from a 
simple button push.  This eliminates the need to build complicated button pads (Although you still could do this) to 
call a sound and have it play into your mixer device.  

The HiFiBerry DAC provides a quality sounding DAC that can filter unecessary noise created by the Pi's built in sound
card.  It also provides two RCA out jacks which can be used with a RCA->unbalanced 1/4 TRS input for your mixer or sound
interface (such as a focusrite 4i4 or similar device).  When a sound is triggered, the sound will play through the DAC into
your mixer or sound interface.  This allows line level mixing and integration with a DAW or OBS without having to go through
your computer.  

## Software

The interface is written in GoLang and aims to be as simple as possible. The web UI provides a simple interface for uploading and managing
sound clips while responding with as low as latency as possibly.  Additionally since it is written in go, you can simply install
the binary, set the binary up to run as a service and you are done.  Connect up your hardware, load up sound clips and you are ready to go.  
After you have uploaded some sounds, configure your Elgato Stream Deck with the URL's provided through the web interface.  

## Limitations 

- You can upload as many songs as you like with any length, however your limit is the space on your SD Card.
- No builtin mechanism exists for backing up your songs and database, but the data is stored on disk and can easily be copied to off to alternative storage.
- You have to configure ALSA to use the HiFiBerry, Please see the documentation for the HiFiBerry or similar DAC you want to use.
- You can use the built in sound device in the raspberry pi for testing, but the sound quality is not super great, but maybe good enough for you.

## Installation

### Raspberry Pi Setup

#### Install packages

```bash
apt-get install mplayer apt-get avahi-utils avahi-daemon pulseaudio-module-zeroconf

```

- mplayer is used by sounder-gong to play sounds.  I did not implement a native golang player, but rather used mplayer instead
- You could use ALSA directly but pulseaudio allows you to play multiple sounds simultaneously.  

#### Configure environment to use pulse audio

Configure ALSA

```bash
sudo vim /etc/asound.conf
```

Add the following contents

```bash
pcm.hifiberry {
  type hw card 1
}

pcm.!default {
  type plug
  slave.pcm "dmixer"
}

pcm.dmixer {
  type dmix
  ipc_key 1024
  slave {
    pcm "hifiberry"
    channels 2
  }
}

ctl.dmixer {
  type hw
  card 1
}
```

Make sure that the raspberry pi has the config overlay for the HiFiBerry DAC

```bash
vim /boot/config.txt
```

**** You will have to reboot the raspberry pi after editing the config.txt ****

Add the following lines are in the general section

```bash
# Enable dac
dtoverlay=hifiberry-dacplus
dtoverlay=i2s-mmap
```

Ensure the avahi-daemon is running

```bash
systemctl install avahi-daemon
systemctl start avahi-daemon
```

Ensure Pulse Audio is running (TODO: Figure out how to initialize this, normally pulseaudio starts with your window manager)

```bash
pulseaudio -D
```

We need to configure PulseAudio to use the right sound card, we use pacmd for this

```bash
pi@raspberrypi:~ $ pacmd
Welcome to PulseAudio 12.2! Use "help" for usage information.
>>> list-sinks
... Lots of stuff printed here, but grab the index of the hifiberry, for me it was #1
>>> set-default-sink 1
>>> exit
```

Start pulseaudio again

```bash
pulseaudio -D
```

Now you should be able to make mplayer play sound through the DAC, please connect
your DAC to something so you can hear the sound!!

```bash
mplayer -ao pulse somesoundfile.wav
```

If you hear the sound, PulseAudio is working correctly

TODO: Find an interface to config






