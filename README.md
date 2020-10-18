# Sounder Gong 

This project aims to build a raspberry pi based sound board that is capable of integrating with a Elgato Stream Deck.

The Elgato Stream Deck is used by pod casters, bloggers and content creators to provide a button pad that integrates
easily with software such as Zoom, OBS and many other types of software to name a few.  

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


