#include <stdint.h>
#include <sys/types.h>
#include <sys/time.h>
#include <stdio.h>
#include <unistd.h>

#include "bcm2835.h"
#include "dht.h"

inline uint64_t get_timestamp(const struct timeval * tv) {
  return tv->tv_sec * 1000000ull + tv->tv_usec;
}

// get_max_counter returns estimated counter value for 100us
uint16_t get_max_counter(int pin) {
  uint64_t before, after;
  struct timeval tv;

  bcm2835_gpio_fsel(pin, BCM2835_GPIO_FSEL_INPT);

  gettimeofday(&tv, NULL);
  before = get_timestamp(&tv);

  for (uint16_t cnt = 0; cnt < 10000; ++cnt) {
    bcm2835_gpio_lev(pin);
  }

  gettimeofday(&tv, NULL);
  after = get_timestamp(&tv);

  return 10000 * 100 / (after - before);
}

void dht_init(struct dht_sensor * dht, int pin) {
  dht->pin = pin;
  dht->max_counter = get_max_counter(pin);
  dht->last_timestamp = 0;
  dht->last_level = 0;
  dht->loglevel = 0;
}

void dht_set_loglevel(struct dht_sensor * dht, int loglevel) {
  dht->loglevel = loglevel;
}

uint64_t read_pulse_width(struct dht_sensor * dht) {
  uint16_t cnt = 0;
  int level = dht->last_level;
  int pin = dht->pin;
  uint64_t timestamp = 0, width = 0;
  struct timeval tv;

  while ((level = bcm2835_gpio_lev(pin)) == dht->last_level) {
    cnt ++;
    if (cnt == dht->max_counter) {
      if (dht->loglevel) fputs(">100us, timed out\n", stderr);
      return 100;
    }
  }
  gettimeofday(&tv, NULL);
  timestamp = get_timestamp(&tv);
  width = timestamp - dht->last_timestamp;
  dht->last_level = level;
  dht->last_timestamp = timestamp;
  return width;
}

int begin_sense(struct dht_sensor * dht) {
  struct timeval tv;
  int pin = dht->pin;
  uint64_t width;

  bcm2835_gpio_fsel(pin, BCM2835_GPIO_FSEL_OUTP);

  if (dht->loglevel) fputs("Initiate read\n", stderr);
  bcm2835_gpio_write(pin, HIGH);
  bcm2835_delay(500);
  bcm2835_gpio_write(pin, LOW);
  bcm2835_delay(20);

  bcm2835_gpio_fsel(pin, BCM2835_GPIO_FSEL_INPT);

  if (dht->loglevel) fputs("Waiting respond\n", stderr);
  while (bcm2835_gpio_lev(pin) == HIGH) {
    bcm2835_delayMicroseconds(1);
  }
  if (dht->loglevel) fputs("Respond received", stderr);
  gettimeofday(&tv, NULL);
  dht->last_timestamp = get_timestamp(&tv);
  dht->last_level = bcm2835_gpio_lev(pin);
  width = read_pulse_width(dht);
  if (dht->loglevel > 1) {
    int level = bcm2835_gpio_lev(dht->pin);
    fprintf(stderr, "Dropping pulse %lluus %s\n", width, level != HIGH ? "HIGH" : " LOW");
  }
  width = read_pulse_width(dht);
  if (dht->loglevel > 1) {
    int level = bcm2835_gpio_lev(dht->pin);
    fprintf(stderr, "Dropping pulse %lluus %s\n", width, level != HIGH ? "HIGH" : " LOW");
  }
  if (width < 70 || dht->last_level != LOW) {
    if (dht->loglevel) fputs("Bad response\n", stderr);
    return 0;
  }
  if (dht->loglevel) fputs("Ready to signal\n", stderr);
  return 1;
}

int dht_read(struct dht_sensor * dht, uint32_t * output) {
  uint8_t data[5] = {0, };

  if(!begin_sense(dht)) return 0;

  if (dht->loglevel) fputs("Read 40bits\n", stderr);
  for(int i = 0; i < 40; i++) {
    uint64_t low = read_pulse_width(dht);
    uint64_t high = read_pulse_width(dht);
    data[i / 8] <<= 1;
    if (low < 30 || low > 70) {
      if (dht->loglevel > 1) {
        fprintf(stderr, "Bad response (LOW) %lluus\n", low);
      }
    } else {
      data[i / 8] |= (high > 50);
    }
    if (dht->loglevel > 1) {
      fprintf(stderr, "#%02d: %3llu %3llu\n", i, low, high);
    }
  }
  if (dht->loglevel > 1) {
    fprintf(stderr, "%02x %02x %02x %02x %02x\n", data[0], data[1], data[2], data[3], data[4]);
  }

  if (data[0] + data[1] + data[2] + data[3] != data[4]) {
    if (dht->loglevel) fputs("Checksum failed\n", stderr);
    return 0;
  }
  *output = (((uint32_t)data[0]) << 24) |
            (((uint32_t)data[1]) << 16) |
            (((uint32_t)data[2]) << 8) |
            ((uint32_t)data[3]);

  if (dht->loglevel) {
    fputs("Succeeded\n", stderr);
    fflush(stderr);
  }
  return 1;
}
