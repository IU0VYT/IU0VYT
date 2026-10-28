/*
 * IU0VYT Flash — prove di confronto con dmrconfig (https://github.com/OpenRTX/dmrconfig)
 * Copyright (C) 2026 Daniele "IU0VYT" Bonifazi
 * SPDX-License-Identifier: GPL-3.0-or-later
 *
 * Sostituisce i collegamenti USB di dmrconfig con un collegamento TCP verso la
 * radio simulata di IU0VYT Flash, così che il codice ORIGINALE di dmrconfig
 * (hid.c, dfu-libusb.c, serial.c e i driver delle radio) parli con la radio
 * simulata. Variabili d'ambiente: DMRCONFIG_HID e DMRCONFIG_DFU = host:porta.
 *
 * HID: ogni scambio è il rapporto di 42 byte inviato (come hid-windows.c:
 * 01 00 lunghezza(2) dati) e il rapporto di 42 byte ricevuto (03 00 lunghezza 00 dati).
 * DFU: per ogni libusb_control_transfer si inviano 8 byte (bmRequestType,
 * bRequest, wValue, wIndex, wLength, little endian) e gli eventuali dati in
 * uscita; si ricevono 4 byte (esito, intero con segno) e i dati in entrata.
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <stdint.h>
#include <netdb.h>
#include <sys/socket.h>
#include <libusb.h>
#include "util.h"

static int tcp_open(const char *env)
{
    const char *a = getenv(env);
    if (!a)
        return -1;
    char host[256];
    strncpy(host, a, sizeof(host) - 1);
    host[sizeof(host) - 1] = 0;
    char *port = strrchr(host, ':');
    if (!port)
        return -1;
    *port++ = 0;
    struct addrinfo hints, *res;
    memset(&hints, 0, sizeof(hints));
    hints.ai_socktype = SOCK_STREAM;
    if (getaddrinfo(host, port, &hints, &res) != 0)
        return -1;
    int fd = socket(res->ai_family, res->ai_socktype, res->ai_protocol);
    if (fd < 0 || connect(fd, res->ai_addr, res->ai_addrlen) < 0) {
        freeaddrinfo(res);
        return -1;
    }
    freeaddrinfo(res);
    return fd;
}

static void xwrite(int fd, const void *b, size_t n)
{
    const unsigned char *p = b;
    while (n > 0) {
        ssize_t k = write(fd, p, n);
        if (k <= 0) {
            fprintf(stderr, "tcpdev: write error\n");
            exit(-1);
        }
        p += k;
        n -= k;
    }
}

static void xread(int fd, void *b, size_t n)
{
    unsigned char *p = b;
    while (n > 0) {
        ssize_t k = read(fd, p, n);
        if (k <= 0) {
            fprintf(stderr, "tcpdev: read error\n");
            exit(-1);
        }
        p += k;
        n -= k;
    }
}

/* ------------------------------------------------------------------ HID */

static int hid_fd = -1;

int hid_init(int vid, int pid)
{
    hid_fd = tcp_open("DMRCONFIG_HID");
    return hid_fd < 0 ? -1 : 0;
}

void hid_close()
{
    if (hid_fd >= 0)
        close(hid_fd);
    hid_fd = -1;
}

void hid_send_recv(const unsigned char *data, unsigned nbytes, unsigned char *rdata, unsigned rlength)
{
    unsigned char buf[42], rbuf[42];
    memset(buf, 0, sizeof(buf));
    buf[0] = 1;
    buf[1] = 0;
    buf[2] = nbytes;
    buf[3] = nbytes >> 8;
    if (nbytes > 0)
        memcpy(buf + 4, data, nbytes);
    xwrite(hid_fd, buf, sizeof(buf));
    xread(hid_fd, rbuf, sizeof(rbuf));
    if (rbuf[0] != 3 || rbuf[1] != 0 || rbuf[3] != 0) {
        fprintf(stderr, "incorrect reply\n");
        exit(-1);
    }
    if (rbuf[2] != rlength) {
        fprintf(stderr, "incorrect reply length %d, expected %d\n", rbuf[2], rlength);
        exit(-1);
    }
    memcpy(rdata, rbuf + 4, rlength);
}

/* ------------------------------------------------------------------ libusb (DFU) */

static int dfu_fd = -1;
static int dummy_ctx, dummy_dev;

int libusb_init(libusb_context **c) { if (c) *c = (libusb_context*)&dummy_ctx; return 0; }
void libusb_exit(libusb_context *c) { }
libusb_device_handle *libusb_open_device_with_vid_pid(libusb_context *c, uint16_t vid, uint16_t pid)
{
    dfu_fd = tcp_open("DMRCONFIG_DFU");
    return dfu_fd < 0 ? NULL : (libusb_device_handle*)&dummy_dev;
}
void libusb_close(libusb_device_handle *h) { if (dfu_fd >= 0) close(dfu_fd); dfu_fd = -1; }
int libusb_kernel_driver_active(libusb_device_handle *h, int i) { return 0; }
int libusb_detach_kernel_driver(libusb_device_handle *h, int i) { return 0; }
int libusb_claim_interface(libusb_device_handle *h, int i) { return 0; }
int libusb_release_interface(libusb_device_handle *h, int i) { return 0; }
const char *libusb_strerror(int e) { return "errore della radio simulata"; }

int libusb_control_transfer(libusb_device_handle *h, uint8_t rt, uint8_t rq, uint16_t value,
    uint16_t index, unsigned char *data, uint16_t length, unsigned int timeout)
{
    unsigned char hdr[8] = { rt, rq, value, value >> 8, index, index >> 8, length, length >> 8 };
    xwrite(dfu_fd, hdr, 8);
    if (!(rt & 0x80) && length > 0)
        xwrite(dfu_fd, data, length);
    unsigned char r[4];
    xread(dfu_fd, r, 4);
    int32_t res = (int32_t)(r[0] | r[1] << 8 | r[2] << 16 | (uint32_t)r[3] << 24);
    if ((rt & 0x80) && res > 0)
        xread(dfu_fd, data, res);
    return res;
}
