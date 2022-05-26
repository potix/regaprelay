# Preparation

## write rasberry pi image to sd card
1. download rasberry pi imager
   - https://www.raspberrypi.com/software/
1. run rasberry pi imager
1. select image
1. setting
   - hostname
   - enable ssh
   - account
   - wifi
   - locale
1. select device
1. write image to sd card

## resize sd card partition (on linux PC)
1. attach sd card
1. run parted
   ```
   parted /dev/<sd card device>
   ```
1. resize partation of sd card
   ```
   resizepart 2 <sd card capacity>GB
   ```
   
## preboot setup (on linux PC)
1. mount sd card
   ```
   mount <sd card device> /mnt
   ```
1. setup files in /mnt/boot
   ```
   touch /mnt/boot/ssh
   vi /mnt/boot/wpa_supplicant.conf
   ---
   ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
   country=JP
   update_config=1
   ap_scan=1

   network={
        ssid="<ssid>"
        psk="<password>"
        scan_ssid=1 
        key_mgmt=WPA_PSK
   }
   ---
   vi /mnt/etc/dhcpcd.conf
   ---
   .
   .
   .
   interface eth0
   static ip_address=<ip address>/24
   static routers=<gateway address>
   static domain_name_servers=<dns addresses>
   ---
   ```
1. setup ssh
   ```
   mkdir /mnt/home/pi/.ssh
   chmod 700 /mnt/home/pi/.ssh
   vi /mnt/home/pi/.ssh/authorized_keys
   ---
   ssh-rsa <public key> foo@bar 
   ---
   chmod 600 /mnt/home/pi/.ssh/authorized_keys
   chmown -R <id 1000 user>:<id 1000 group> /mnt/home/pi/.ssh
   vi /mnt/etc/ssh/sshd_config
   ---
   .
   .
   .
   #PermitRootLogin prohibit-password
   PermitRootLogin no
   .
   .
   .
   #PasswordAuthentication yes
   PasswordAuthentication no
   .
   .
   .
   ---
   ``` 

## boot rasberry pi
1. attach sd card to rasberry pi 
1. power on
1. login to rasberry pi
1. apt update


# Setup

## install packages   
   - golang
     - I build from source code in my case.


## usb gadget mode (USB OTG) setting
   ```
   sudo vi /boot/config.txt
   ---
   .
   .
   .
   # Uncomment this to enable infrared communication.
   #dtoverlay=gpio-ir,gpio_pin=17
   #dtoverlay=gpio-ir-tx,gpio_pin=18
   dtoverlay=dwc2
   .
   .
   .
   ---
   echo "dwc2" | sudo tee -a /etc/modules
　 echo "libcomposite" | sudo tee -a /etc/modules
   sudo reboot
   ```
