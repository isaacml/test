/*
Navicat SQLite Data Transfer

Source Server         : Download DB
Source Server Version : 30808
Source Host           : :0

Target Server Type    : SQLite
Target Server Version : 30808
File Encoding         : 65001

Date: 2016-05-10 17:24:01
*/

PRAGMA foreign_keys = OFF;

-- ----------------------------
-- Table structure for segmentos
-- ----------------------------
DROP TABLE IF EXISTS "main"."segmentos";
CREATE TABLE "segmentos" (
"filename"  TEXT(255),
"bytes"  INTEGER,
"md5sum"  TEXT(32),
"fvideo"  TEXT(7),
"faudio"  TEXT(7),
"hres"  INTEGER,
"vres"  INTEGER,
"num_fps"  INTEGER,
"den_fps"  INTEGER,
"vbitrate"  INTEGER,
"abitrate"  INTEGER,
"block"  TEXT,
"next"  TEXT,
"semaforo"  TEXT(1),
"duration"  INTEGER,
"timestamp"  INTEGER,
"mac"  TEXT(16),
"last_connect"  INTEGER,
"tv_id"  INTEGER NOT NULL
);

-- ----------------------------
-- Table structure for sqlite_sequence
-- ----------------------------
DROP TABLE IF EXISTS "main"."sqlite_sequence";
CREATE TABLE sqlite_sequence(name,seq);
