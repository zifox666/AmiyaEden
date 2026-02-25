# SDE - EVE 数据资产管理模块

## 概述

SDE 模块负责更新和管理 EVE Online 的数据资产（Static Data Export）。这些数据包括游戏中的物品、角色、技能等静态信息。通过定期从官方 SDE 数据源获取最新数据，确保系统中的数据保持最新和准确。并为其他模块提供数据查询接口

## 功能

1. SDE更新

    ```markdown
    # EVE SDE Converter

    A CLI tool built with TypeScript to convert EVE Online Static Data Export (SDE) from JSONL format to MySQL dump and SQLite database files. The project is designed to run in GitHub Actions.

    ## Features

    - Fetches the latest EVE Online SDE build number from the official API
    - Downloads and extracts the JSONL data archive
    - Processes JSONL files and maps data to MySQL schema
    - Generates MySQL dump files
    - Converts MySQL dump to SQLite format using the included mysql2sqlite utility
    - Supports incremental updates by checking for existing build tags in GitHub Actions
    ```

    `https://github.com/zifox666/eve-sde-converter/releases` 更新地址在这里 每天20：00检查更新
    项目提供的sql表[sde.sql](./sdel.sql) 需要支持在网页手动更新以及查看sde版本功能
    从relea获取数据后 解压sql文件 导入到数据库中 之后可以删除sql文件 需要记住sde版本号

2. 数据查询接口

    - 提供函数API和网页API供其他模块查询SDE数据，网页API需要接入API KEY权限控制

## 需要的查询接口

1. 物品翻译接口

    - ttrnTranslations
    - tc_id: 暂时用的到的 TC_TYPES_ID = 8 TC_GROUP_ID = 7 TC_CATEGORY_ID = 6
    - 通过输入tc_id,key_id,language_id查询翻译结果

2. 名称模糊查询接口

    - 可能需要使用jieba对中文分词

3. type_id查询接口

    - 返回type_id的详细信息 包括group_id category_id等
