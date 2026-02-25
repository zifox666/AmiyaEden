-- MySQL dump 10.19  Distrib 10.3.38-MariaDB, for Linux (x86_64)
--
-- Host: localhost    Database: sdeyaml
-- ------------------------------------------------------
-- Server version	10.3.38-MariaDB

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `agtAgentTypes`
--

DROP TABLE IF EXISTS `agtAgentTypes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `agtAgentTypes` (
  `agentTypeID` int(11) NOT NULL,
  `agentType` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`agentTypeID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `agtAgentTypes`
--

LOCK TABLES `agtAgentTypes` WRITE;
/*!40000 ALTER TABLE `agtAgentTypes` DISABLE KEYS */;
/*!40000 ALTER TABLE `agtAgentTypes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `agtAgents`
--

DROP TABLE IF EXISTS `agtAgents`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `agtAgents` (
  `agentID` int(11) NOT NULL,
  `divisionID` int(11) DEFAULT NULL,
  `corporationID` int(11) DEFAULT NULL,
  `locationID` int(11) DEFAULT NULL,
  `level` int(11) DEFAULT NULL,
  `quality` int(11) DEFAULT NULL,
  `agentTypeID` int(11) DEFAULT NULL,
  `isLocator` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`agentID`),
  KEY `ix_agtAgents_corporationID` (`corporationID`),
  KEY `ix_agtAgents_locationID` (`locationID`),
  CONSTRAINT `aa_isloc` CHECK (`isLocator` in (0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `agtAgents`
--

LOCK TABLES `agtAgents` WRITE;
/*!40000 ALTER TABLE `agtAgents` DISABLE KEYS */;
/*!40000 ALTER TABLE `agtAgents` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `agtAgentsInSpace`
--

DROP TABLE IF EXISTS `agtAgentsInSpace`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `agtAgentsInSpace` (
  `agentID` int(11) NOT NULL,
  `dungeonID` int(11) DEFAULT NULL,
  `solarSystemID` int(11) DEFAULT NULL,
  `spawnPointID` int(11) DEFAULT NULL,
  `typeID` int(11) DEFAULT NULL,
  PRIMARY KEY (`agentID`),
  KEY `ix_agtAgentsInSpace_solarSystemID` (`solarSystemID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `agtAgentsInSpace`
--

LOCK TABLES `agtAgentsInSpace` WRITE;
/*!40000 ALTER TABLE `agtAgentsInSpace` DISABLE KEYS */;
/*!40000 ALTER TABLE `agtAgentsInSpace` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `agtResearchAgents`
--

DROP TABLE IF EXISTS `agtResearchAgents`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `agtResearchAgents` (
  `agentID` int(11) NOT NULL,
  `typeID` int(11) NOT NULL,
  PRIMARY KEY (`agentID`,`typeID`),
  KEY `ix_agtResearchAgents_typeID` (`typeID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `agtResearchAgents`
--

LOCK TABLES `agtResearchAgents` WRITE;
/*!40000 ALTER TABLE `agtResearchAgents` DISABLE KEYS */;
/*!40000 ALTER TABLE `agtResearchAgents` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `certCerts`
--

DROP TABLE IF EXISTS `certCerts`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `certCerts` (
  `certID` int(11) NOT NULL,
  `description` text DEFAULT NULL,
  `groupID` int(11) DEFAULT NULL,
  `name` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`certID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `certCerts`
--

LOCK TABLES `certCerts` WRITE;
/*!40000 ALTER TABLE `certCerts` DISABLE KEYS */;
/*!40000 ALTER TABLE `certCerts` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `certMasteries`
--

DROP TABLE IF EXISTS `certMasteries`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `certMasteries` (
  `typeID` int(11) DEFAULT NULL,
  `masteryLevel` int(11) DEFAULT NULL,
  `certID` int(11) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `certMasteries`
--

LOCK TABLES `certMasteries` WRITE;
/*!40000 ALTER TABLE `certMasteries` DISABLE KEYS */;
/*!40000 ALTER TABLE `certMasteries` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `certSkills`
--

DROP TABLE IF EXISTS `certSkills`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `certSkills` (
  `certID` int(11) DEFAULT NULL,
  `skillID` int(11) DEFAULT NULL,
  `certLevelInt` int(11) DEFAULT NULL,
  `skillLevel` int(11) DEFAULT NULL,
  `certLevelText` varchar(8) DEFAULT NULL,
  KEY `ix_certSkills_skillID` (`skillID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `certSkills`
--

LOCK TABLES `certSkills` WRITE;
/*!40000 ALTER TABLE `certSkills` DISABLE KEYS */;
/*!40000 ALTER TABLE `certSkills` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `charFactions`
--
-- 

-- 
--
-- Table structure for table `chrAncestries`
--

DROP TABLE IF EXISTS `chrAncestries`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `chrAncestries` (
  `ancestryID` int(11) NOT NULL,
  `ancestryName` varchar(100) DEFAULT NULL,
  `bloodlineID` int(11) DEFAULT NULL,
  `description` varchar(1000) DEFAULT NULL,
  `perception` int(11) DEFAULT NULL,
  `willpower` int(11) DEFAULT NULL,
  `charisma` int(11) DEFAULT NULL,
  `memory` int(11) DEFAULT NULL,
  `intelligence` int(11) DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  `shortDescription` varchar(500) DEFAULT NULL,
  PRIMARY KEY (`ancestryID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `chrAncestries`
--

LOCK TABLES `chrAncestries` WRITE;
/*!40000 ALTER TABLE `chrAncestries` DISABLE KEYS */;
/*!40000 ALTER TABLE `chrAncestries` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `chrAttributes`
--

DROP TABLE IF EXISTS `chrAttributes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `chrAttributes` (
  `attributeID` int(11) NOT NULL,
  `attributeName` varchar(100) DEFAULT NULL,
  `description` varchar(1000) DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  `shortDescription` varchar(500) DEFAULT NULL,
  `notes` varchar(500) DEFAULT NULL,
  PRIMARY KEY (`attributeID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `chrAttributes`
--

LOCK TABLES `chrAttributes` WRITE;
/*!40000 ALTER TABLE `chrAttributes` DISABLE KEYS */;
/*!40000 ALTER TABLE `chrAttributes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `chrBloodlines`
--

DROP TABLE IF EXISTS `chrBloodlines`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `chrBloodlines` (
  `bloodlineID` int(11) NOT NULL,
  `bloodlineName` varchar(100) DEFAULT NULL,
  `raceID` int(11) DEFAULT NULL,
  `description` varchar(1000) DEFAULT NULL,
  `maleDescription` varchar(1000) DEFAULT NULL,
  `femaleDescription` varchar(1000) DEFAULT NULL,
  `shipTypeID` int(11) DEFAULT NULL,
  `corporationID` int(11) DEFAULT NULL,
  `perception` int(11) DEFAULT NULL,
  `willpower` int(11) DEFAULT NULL,
  `charisma` int(11) DEFAULT NULL,
  `memory` int(11) DEFAULT NULL,
  `intelligence` int(11) DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  `shortDescription` varchar(500) DEFAULT NULL,
  `shortMaleDescription` varchar(500) DEFAULT NULL,
  `shortFemaleDescription` varchar(500) DEFAULT NULL,
  PRIMARY KEY (`bloodlineID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `chrBloodlines`
--

LOCK TABLES `chrBloodlines` WRITE;
/*!40000 ALTER TABLE `chrBloodlines` DISABLE KEYS */;
/*!40000 ALTER TABLE `chrBloodlines` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `chrFactions`
--

DROP TABLE IF EXISTS `chrFactions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `chrFactions` (
  `factionID` int(11) NOT NULL,
  `factionName` varchar(100) DEFAULT NULL,
  `description` varchar(2000) DEFAULT NULL,
  `raceIDs` int(11) DEFAULT NULL,
  `solarSystemID` int(11) DEFAULT NULL,
  `corporationID` int(11) DEFAULT NULL,
  `sizeFactor` float DEFAULT NULL,
  `stationCount` int(11) DEFAULT NULL,
  `stationSystemCount` int(11) DEFAULT NULL,
  `militiaCorporationID` int(11) DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  PRIMARY KEY (`factionID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `chrFactions`
--

LOCK TABLES `chrFactions` WRITE;
/*!40000 ALTER TABLE `chrFactions` DISABLE KEYS */;
/*!40000 ALTER TABLE `chrFactions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `chrRaces`
--

DROP TABLE IF EXISTS `chrRaces`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `chrRaces` (
  `raceID` int(11) NOT NULL,
  `raceName` varchar(100) DEFAULT NULL,
  `description` varchar(1000) DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  `shortDescription` varchar(500) DEFAULT NULL,
  PRIMARY KEY (`raceID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `chrRaces`
--

LOCK TABLES `chrRaces` WRITE;
/*!40000 ALTER TABLE `chrRaces` DISABLE KEYS */;
/*!40000 ALTER TABLE `chrRaces` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `crpActivities`
--

DROP TABLE IF EXISTS `crpActivities`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `crpActivities` (
  `activityID` int(11) NOT NULL,
  `activityName` varchar(100) DEFAULT NULL,
  `description` varchar(1000) DEFAULT NULL,
  PRIMARY KEY (`activityID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `crpActivities`
--

LOCK TABLES `crpActivities` WRITE;
/*!40000 ALTER TABLE `crpActivities` DISABLE KEYS */;
/*!40000 ALTER TABLE `crpActivities` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `crpNPCCorporationDivisions`
--
-- 

-- 
--
-- Table structure for table `crpNPCCorporationResearchFields`
--
-- 

-- 
--
-- Table structure for table `crpNPCCorporationTrades`
--
-- 

-- 
--
-- Table structure for table `crpNPCCorporations`
--

DROP TABLE IF EXISTS `crpNPCCorporations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `crpNPCCorporations` (
  `corporationID` int(11) NOT NULL,
  `size` char(1) DEFAULT NULL,
  `extent` char(1) DEFAULT NULL,
  `solarSystemID` int(11) DEFAULT NULL,
  `investorID1` int(11) DEFAULT NULL,
  `investorShares1` int(11) DEFAULT NULL,
  `investorID2` int(11) DEFAULT NULL,
  `investorShares2` int(11) DEFAULT NULL,
  `investorID3` int(11) DEFAULT NULL,
  `investorShares3` int(11) DEFAULT NULL,
  `investorID4` int(11) DEFAULT NULL,
  `investorShares4` int(11) DEFAULT NULL,
  `friendID` int(11) DEFAULT NULL,
  `enemyID` int(11) DEFAULT NULL,
  `publicShares` int(11) DEFAULT NULL,
  `initialPrice` int(11) DEFAULT NULL,
  `minSecurity` float DEFAULT NULL,
  `scattered` tinyint(1) DEFAULT NULL,
  `fringe` int(11) DEFAULT NULL,
  `corridor` int(11) DEFAULT NULL,
  `hub` int(11) DEFAULT NULL,
  `border` int(11) DEFAULT NULL,
  `factionID` int(11) DEFAULT NULL,
  `sizeFactor` float DEFAULT NULL,
  `stationCount` int(11) DEFAULT NULL,
  `stationSystemCount` int(11) DEFAULT NULL,
  `description` varchar(4000) DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  PRIMARY KEY (`corporationID`),
  CONSTRAINT `cnpcc_scatt` CHECK (`scattered` in (0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `crpNPCCorporations`
--

LOCK TABLES `crpNPCCorporations` WRITE;
/*!40000 ALTER TABLE `crpNPCCorporations` DISABLE KEYS */;
/*!40000 ALTER TABLE `crpNPCCorporations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `crpNPCDivisions`
--
-- 

-- 
--
-- Table structure for table `dgmAttributeCategories`
--

DROP TABLE IF EXISTS `dgmAttributeCategories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dgmAttributeCategories` (
  `categoryID` int(11) NOT NULL,
  `categoryName` varchar(50) DEFAULT NULL,
  `categoryDescription` varchar(200) DEFAULT NULL,
  PRIMARY KEY (`categoryID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `dgmAttributeCategories`
--

LOCK TABLES `dgmAttributeCategories` WRITE;
/*!40000 ALTER TABLE `dgmAttributeCategories` DISABLE KEYS */;
/*!40000 ALTER TABLE `dgmAttributeCategories` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `dgmAttributeTypes`
--

DROP TABLE IF EXISTS `dgmAttributeTypes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dgmAttributeTypes` (
  `attributeID` int(11) NOT NULL,
  `attributeName` varchar(100) DEFAULT NULL,
  `description` varchar(1000) DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  `defaultValue` float DEFAULT NULL,
  `published` tinyint(1) DEFAULT NULL,
  `displayName` varchar(150) DEFAULT NULL,
  `unitID` int(11) DEFAULT NULL,
  `stackable` tinyint(1) DEFAULT NULL,
  `highIsGood` tinyint(1) DEFAULT NULL,
  `categoryID` int(11) DEFAULT NULL,
  PRIMARY KEY (`attributeID`),
  CONSTRAINT `dat_pub` CHECK (`published` in (0,1)),
  CONSTRAINT `dat_stack` CHECK (`stackable` in (0,1)),
  CONSTRAINT `dat_hig` CHECK (`highIsGood` in (0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `dgmAttributeTypes`
--

LOCK TABLES `dgmAttributeTypes` WRITE;
/*!40000 ALTER TABLE `dgmAttributeTypes` DISABLE KEYS */;
/*!40000 ALTER TABLE `dgmAttributeTypes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `dgmEffects`
--

DROP TABLE IF EXISTS `dgmEffects`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dgmEffects` (
  `effectID` int(11) NOT NULL,
  `effectName` varchar(400) DEFAULT NULL,
  `effectCategory` int(11) DEFAULT NULL,
  `preExpression` int(11) DEFAULT NULL,
  `postExpression` int(11) DEFAULT NULL,
  `description` varchar(1000) DEFAULT NULL,
  `guid` varchar(60) DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  `isOffensive` tinyint(1) DEFAULT NULL,
  `isAssistance` tinyint(1) DEFAULT NULL,
  `durationAttributeID` int(11) DEFAULT NULL,
  `trackingSpeedAttributeID` int(11) DEFAULT NULL,
  `dischargeAttributeID` int(11) DEFAULT NULL,
  `rangeAttributeID` int(11) DEFAULT NULL,
  `falloffAttributeID` int(11) DEFAULT NULL,
  `disallowAutoRepeat` tinyint(1) DEFAULT NULL,
  `published` tinyint(1) DEFAULT NULL,
  `displayName` varchar(100) DEFAULT NULL,
  `isWarpSafe` tinyint(1) DEFAULT NULL,
  `rangeChance` tinyint(1) DEFAULT NULL,
  `electronicChance` tinyint(1) DEFAULT NULL,
  `propulsionChance` tinyint(1) DEFAULT NULL,
  `distribution` int(11) DEFAULT NULL,
  `sfxName` varchar(20) DEFAULT NULL,
  `npcUsageChanceAttributeID` int(11) DEFAULT NULL,
  `npcActivationChanceAttributeID` int(11) DEFAULT NULL,
  `fittingUsageChanceAttributeID` int(11) DEFAULT NULL,
  `modifierInfo` text DEFAULT NULL,
  PRIMARY KEY (`effectID`),
  CONSTRAINT `de_offense` CHECK (`isOffensive` in (0,1)),
  CONSTRAINT `de_assist` CHECK (`isAssistance` in (0,1)),
  CONSTRAINT `de_disallowar` CHECK (`disallowAutoRepeat` in (0,1)),
  CONSTRAINT `de_published` CHECK (`published` in (0,1)),
  CONSTRAINT `de_warpsafe` CHECK (`isWarpSafe` in (0,1)),
  CONSTRAINT `de_rangechance` CHECK (`rangeChance` in (0,1)),
  CONSTRAINT `de_elecchance` CHECK (`electronicChance` in (0,1)),
  CONSTRAINT `de_propchance` CHECK (`propulsionChance` in (0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `dgmEffects`
--

LOCK TABLES `dgmEffects` WRITE;
/*!40000 ALTER TABLE `dgmEffects` DISABLE KEYS */;
/*!40000 ALTER TABLE `dgmEffects` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `dgmExpressions`
--
-- 

-- 
--
-- Table structure for table `dgmTypeAttributes`
--

DROP TABLE IF EXISTS `dgmTypeAttributes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dgmTypeAttributes` (
  `typeID` int(11) NOT NULL,
  `attributeID` int(11) NOT NULL,
  `valueInt` int(11) DEFAULT NULL,
  `valueFloat` float DEFAULT NULL,
  PRIMARY KEY (`typeID`,`attributeID`),
  KEY `ix_dgmTypeAttributes_attributeID` (`attributeID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `dgmTypeAttributes`
--

LOCK TABLES `dgmTypeAttributes` WRITE;
/*!40000 ALTER TABLE `dgmTypeAttributes` DISABLE KEYS */;
/*!40000 ALTER TABLE `dgmTypeAttributes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `dgmTypeEffects`
--

DROP TABLE IF EXISTS `dgmTypeEffects`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `dgmTypeEffects` (
  `typeID` int(11) NOT NULL,
  `effectID` int(11) NOT NULL,
  `isDefault` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`typeID`,`effectID`),
  CONSTRAINT `dte_default` CHECK (`isDefault` in (0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `dgmTypeEffects`
--

LOCK TABLES `dgmTypeEffects` WRITE;
/*!40000 ALTER TABLE `dgmTypeEffects` DISABLE KEYS */;
/*!40000 ALTER TABLE `dgmTypeEffects` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `eveGraphics`
--

DROP TABLE IF EXISTS `eveGraphics`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `eveGraphics` (
  `graphicID` int(11) NOT NULL,
  `sofFactionName` varchar(100) DEFAULT NULL,
  `graphicFile` varchar(256) DEFAULT NULL,
  `sofHullName` varchar(100) DEFAULT NULL,
  `sofRaceName` varchar(100) DEFAULT NULL,
  `description` text DEFAULT NULL,
  PRIMARY KEY (`graphicID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `eveGraphics`
--

LOCK TABLES `eveGraphics` WRITE;
/*!40000 ALTER TABLE `eveGraphics` DISABLE KEYS */;
/*!40000 ALTER TABLE `eveGraphics` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `eveIcons`
--

DROP TABLE IF EXISTS `eveIcons`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `eveIcons` (
  `iconID` int(11) NOT NULL,
  `iconFile` varchar(500) DEFAULT NULL,
  `description` text DEFAULT NULL,
  PRIMARY KEY (`iconID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `eveIcons`
--

LOCK TABLES `eveIcons` WRITE;
/*!40000 ALTER TABLE `eveIcons` DISABLE KEYS */;
/*!40000 ALTER TABLE `eveIcons` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `eveUnits`
--

DROP TABLE IF EXISTS `eveUnits`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `eveUnits` (
  `unitID` int(11) NOT NULL,
  `unitName` varchar(100) DEFAULT NULL,
  `displayName` varchar(50) DEFAULT NULL,
  `description` varchar(1000) DEFAULT NULL,
  PRIMARY KEY (`unitID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `eveUnits`
--

LOCK TABLES `eveUnits` WRITE;
/*!40000 ALTER TABLE `eveUnits` DISABLE KEYS */;
/*!40000 ALTER TABLE `eveUnits` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `industryActivity`
--
-- 

-- 
--
-- Table structure for table `industryActivityMaterials`
--
-- 

-- 
--
-- Table structure for table `industryActivityProbabilities`
--
-- 

-- 
--
-- Table structure for table `industryActivityProducts`
--
-- 

-- 
--
-- Table structure for table `industryActivityRaces`
--
-- 

-- 
--
-- Table structure for table `industryActivitySkills`
--
-- 

-- 
--
-- Table structure for table `industryBlueprints`
--

DROP TABLE IF EXISTS `industryBlueprints`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `industryBlueprints` (
  `typeID` int(11) NOT NULL,
  `maxProductionLimit` int(11) DEFAULT NULL,
  PRIMARY KEY (`typeID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `industryBlueprints`
--

LOCK TABLES `industryBlueprints` WRITE;
/*!40000 ALTER TABLE `industryBlueprints` DISABLE KEYS */;
/*!40000 ALTER TABLE `industryBlueprints` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `invCategories`
--

DROP TABLE IF EXISTS `invCategories`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `invCategories` (
  `categoryID` int(11) NOT NULL,
  `categoryName` varchar(100) DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  `published` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`categoryID`),
  CONSTRAINT `invcat_published` CHECK (`published` in (0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invCategories`
--

LOCK TABLES `invCategories` WRITE;
/*!40000 ALTER TABLE `invCategories` DISABLE KEYS */;
/*!40000 ALTER TABLE `invCategories` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `invContrabandTypes`
--

DROP TABLE IF EXISTS `invContrabandTypes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `invContrabandTypes` (
  `factionID` int(11) NOT NULL,
  `typeID` int(11) NOT NULL,
  `standingLoss` double DEFAULT NULL,
  `confiscateMinSec` double DEFAULT NULL,
  `fineByValue` double DEFAULT NULL,
  `attackMinSec` double DEFAULT NULL,
  PRIMARY KEY (`factionID`,`typeID`),
  KEY `ix_invContrabandTypes_typeID` (`typeID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invContrabandTypes`
--

LOCK TABLES `invContrabandTypes` WRITE;
/*!40000 ALTER TABLE `invContrabandTypes` DISABLE KEYS */;
/*!40000 ALTER TABLE `invContrabandTypes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `invControlTowerResourcePurposes`
--
-- 

-- 
--
-- Table structure for table `invControlTowerResources`
--

DROP TABLE IF EXISTS `invControlTowerResources`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `invControlTowerResources` (
  `controlTowerTypeID` int(11) NOT NULL,
  `resourceTypeID` int(11) NOT NULL,
  `purpose` int(11) DEFAULT NULL,
  `quantity` int(11) DEFAULT NULL,
  `minSecurityLevel` double DEFAULT NULL,
  `factionID` int(11) DEFAULT NULL,
  PRIMARY KEY (`controlTowerTypeID`,`resourceTypeID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invControlTowerResources`
--

LOCK TABLES `invControlTowerResources` WRITE;
/*!40000 ALTER TABLE `invControlTowerResources` DISABLE KEYS */;
/*!40000 ALTER TABLE `invControlTowerResources` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `invFlags`
--
-- 

DROP TABLE IF EXISTS `invFlags`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `invFlags` (
  `flagID` int(11) NOT NULL,
  `flagName` varchar(200) DEFAULT NULL,
  `flagText` varchar(100) DEFAULT NULL,
  `orderID` int(11) DEFAULT NULL,
  PRIMARY KEY (`flagID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invFlags`
--

LOCK TABLES `invFlags` WRITE;
/*!40000 ALTER TABLE `invFlags` DISABLE KEYS */;
INSERT INTO `invFlags` VALUES (0,'None','None',0),(1,'Wallet','Wallet',10),(2,'Offices','OfficeFolder',0),(3,'Wardrobe','Wardrobe',0),(4,'Hangar','Hangar',30),(5,'Cargo','Cargo',3000),(6,'OfficeImpound','Impounded Offices',0),(7,'Skill','Skill',15),(8,'Reward','Reward',17),(11,'LoSlot0','Low power slot 1',0),(12,'LoSlot1','Low power slot 2',0),(13,'LoSlot2','Low power slot 3',0),(14,'LoSlot3','Low power slot 4',0),(15,'LoSlot4','Low power slot 5',0),(16,'LoSlot5','Low power slot 6',0),(17,'LoSlot6','Low power slot 7',0),(18,'LoSlot7','Low power slot 8',0),(19,'MedSlot0','Medium power slot 1',0),(20,'MedSlot1','Medium power slot 2',0),(21,'MedSlot2','Medium power slot 3',0),(22,'MedSlot3','Medium power slot 4',0),(23,'MedSlot4','Medium power slot 5',0),(24,'MedSlot5','Medium power slot 6',0),(25,'MedSlot6','Medium power slot 7',0),(26,'MedSlot7','Medium power slot 8',0),(27,'HiSlot0','High power slot 1',0),(28,'HiSlot1','High power slot 2',0),(29,'HiSlot2','High power slot 3',0),(30,'HiSlot3','High power slot 4',0),(31,'HiSlot4','High power slot 5',0),(32,'HiSlot5','High power slot 6',0),(33,'HiSlot6','High power slot 7',0),(34,'HiSlot7','High power slot 8',0),(35,'Fixed Slot','Fixed Slot',0),(36,'AssetSafety','Asset Safety',0),(56,'Capsule','Capsule',0),(57,'Pilot','Pilot',0),(61,'Skill In Training','Skill in training',0),(62,'CorpMarket','Corporation Market Deliveries / Returns',0),(63,'Locked','Locked item, can not be moved unless unlocked',0),(64,'Unlocked','Unlocked item, can be moved',0),(70,'Office Slot 1','Office slot 1',0),(71,'Office Slot 2','Office slot 2',0),(72,'Office Slot 3','Office slot 3',0),(73,'Office Slot 4','Office slot 4',0),(74,'Office Slot 5','Office slot 5',0),(75,'Office Slot 6','Office slot 6',0),(76,'Office Slot 7','Office slot 7',0),(77,'Office Slot 8','Office slot 8',0),(78,'Office Slot 9','Office slot 9',0),(79,'Office Slot 10','Office slot 10',0),(80,'Office Slot 11','Office slot 11',0),(81,'Office Slot 12','Office slot 12',0),(82,'Office Slot 13','Office slot 13',0),(83,'Office Slot 14','Office slot 14',0),(84,'Office Slot 15','Office slot 15',0),(85,'Office Slot 16','Office slot 16',0),(86,'Bonus','Bonus',0),(87,'DroneBay','Drone Bay',0),(88,'Booster','Booster',0),(89,'Implant','Implant',0),(90,'ShipHangar','Ship Hangar',0),(91,'ShipOffline','Ship Offline',0),(92,'RigSlot0','Rig power slot 1',0),(93,'RigSlot1','Rig power slot 2',0),(94,'RigSlot2','Rig power slot 3',0),(95,'RigSlot3','Rig power slot 4',0),(96,'RigSlot4','Rig power slot 5',0),(97,'RigSlot5','Rig power slot 6',0),(98,'RigSlot6','Rig power slot 7',0),(99,'RigSlot7','Rig power slot 8',0),(115,'CorpSAG1','Corp Security Access Group 1',0),(116,'CorpSAG2','Corp Security Access Group 2',0),(117,'CorpSAG3','Corp Security Access Group 3',0),(118,'CorpSAG4','Corp Security Access Group 4',0),(119,'CorpSAG5','Corp Security Access Group 5',0),(120,'CorpSAG6','Corp Security Access Group 6',0),(121,'CorpSAG7','Corp Security Access Group 7',0),(122,'SecondaryStorage','Secondary Storage',0),(125,'SubSystem0','Sub system slot 0',0),(126,'SubSystem1','Sub system slot 1',0),(127,'SubSystem2','Sub system slot 2',0),(128,'SubSystem3','Sub system slot 3',0),(129,'SubSystem4','Sub system slot 4',0),(130,'SubSystem5','Sub system slot 5',0),(131,'SubSystem6','Sub system slot 6',0),(132,'SubSystem7','Sub system slot 7',0),(133,'SpecializedFuelBay','Specialized Fuel Bay',0),(134,'SpecializedAsteroidHold','Specialized Asteroid Hold',0),(135,'SpecializedGasHold','Specialized Gas Hold',0),(136,'SpecializedMineralHold','Specialized Mineral Hold',0),(137,'SpecializedSalvageHold','Specialized Salvage Hold',0),(138,'SpecializedShipHold','Specialized Ship Hold',0),(139,'SpecializedSmallShipHold','Specialized Small Ship Hold',0),(140,'SpecializedMediumShipHold','Specialized Medium Ship Hold',0),(141,'SpecializedLargeShipHold','Specialized Large Ship Hold',0),(142,'SpecializedIndustrialShipHold','Specialized Industrial Ship Hold',0),(143,'SpecializedAmmoHold','Specialized Ammo Hold',0),(144,'StructureActive','Structure Active',0),(145,'StructureInactive','Structure Inactive',0),(146,'JunkyardReprocessed','This item was put into a junkyard through reprocession.',0),(147,'JunkyardTrashed','This item was put into a junkyard through being trashed by its owner.',0),(148,'SpecializedCommandCenterHold','Specialized Command Center Hold',0),(149,'SpecializedPlanetaryCommoditiesHold','Specialized Planetary Commodities Hold',0),(150,'PlanetSurface','Planet Surface',0),(151,'SpecializedMaterialBay','Specialized Material Bay',0),(152,'DustCharacterDatabank','Dust Character Databank',0),(153,'DustCharacterBattle','Dust Character Battle',0),(154,'QuafeBay','Quafe Bay',0),(155,'FleetHangar','Fleet Hangar',0),(156,'HiddenModifiers','Hidden Modifiers',0),(157,'StructureOffline','Structure Offline',0),(158,'FighterBay','Fighter Bay',0),(159,'FighterTube0','Fighter Tube 0',0),(160,'FighterTube1','Fighter Tube 1',0),(161,'FighterTube2','Fighter Tube 2',0),(162,'FighterTube3','Fighter Tube 3',0),(163,'FighterTube4','Fighter Tube 4',0),(164,'StructureServiceSlot0','Structure service slot 1',0),(165,'StructureServiceSlot1','Structure service slot 2',0),(166,'StructureServiceSlot2','Structure service slot 3',0),(167,'StructureServiceSlot3','Structure service slot 4',0),(168,'StructureServiceSlot4','Structure service slot 5',0),(169,'StructureServiceSlot5','Structure service slot 6',0),(170,'StructureServiceSlot6','Structure service slot 7',0),(171,'StructureServiceSlot7','Structure service slot 8',0),(172,'StructureFuel','Structure Fuel',0),(173,'Deliveries','Deliveries',0),(174,'CrateLoot','Crate Loot',0),(176,'BoosterBay','Booster Hold',0),(177,'SubsystemBay','Subsystem Hold',0),(178,'Raffles','Raffles Hangar',0),(179,'FrigateEscapeBay','Frigate escape bay Hangar',0),(180,'StructureDeedBay','Structure Deed Bay',0),(181,'SpecializedIceHold','Specialized Ice Hold',0),(182,'SpecializedAsteroidHold','Specialized Asteroid Hold',0),(183,'MobileDepot','Mobile Depot',0),(184,'CorpProjectsHangar','Corporation Projects Hangar ',0),(185,'ColonyResourcesHold','Infrastructure Hold',0),(186,'MoonMaterialBay','Moon Material Bay',0),(187,'CapsuleerDeliveries','Capsuleer Deliveries',0);
/*!40000 ALTER TABLE `invFlags` ENABLE KEYS */;
UNLOCK TABLES;

-- 
--
-- Table structure for table `invGroups`
--

DROP TABLE IF EXISTS `invGroups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `invGroups` (
  `groupID` int(11) NOT NULL,
  `categoryID` int(11) DEFAULT NULL,
  `groupName` varchar(100) DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  `useBasePrice` tinyint(1) DEFAULT NULL,
  `anchored` tinyint(1) DEFAULT NULL,
  `anchorable` tinyint(1) DEFAULT NULL,
  `fittableNonSingleton` tinyint(1) DEFAULT NULL,
  `published` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`groupID`),
  KEY `ix_invGroups_categoryID` (`categoryID`),
  CONSTRAINT `invgroup_usebaseprice` CHECK (`useBasePrice` in (0,1)),
  CONSTRAINT `invgroup_anchored` CHECK (`anchored` in (0,1)),
  CONSTRAINT `invgroup_anchorable` CHECK (`anchorable` in (0,1)),
  CONSTRAINT `invgroup_fitnonsingle` CHECK (`fittableNonSingleton` in (0,1)),
  CONSTRAINT `invgroup_published` CHECK (`published` in (0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invGroups`
--

LOCK TABLES `invGroups` WRITE;
/*!40000 ALTER TABLE `invGroups` DISABLE KEYS */;
/*!40000 ALTER TABLE `invGroups` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `invItems`
--
-- 

-- 
--
-- Table structure for table `invMarketGroups`
--

DROP TABLE IF EXISTS `invMarketGroups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `invMarketGroups` (
  `marketGroupID` int(11) NOT NULL,
  `parentGroupID` int(11) DEFAULT NULL,
  `marketGroupName` varchar(100) DEFAULT NULL,
  `description` varchar(3000) DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  `hasTypes` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`marketGroupID`),
  CONSTRAINT `invmarketgroups_hastypes` CHECK (`hasTypes` in (0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invMarketGroups`
--

LOCK TABLES `invMarketGroups` WRITE;
/*!40000 ALTER TABLE `invMarketGroups` DISABLE KEYS */;
/*!40000 ALTER TABLE `invMarketGroups` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `invMetaGroups`
--

DROP TABLE IF EXISTS `invMetaGroups`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `invMetaGroups` (
  `metaGroupID` int(11) NOT NULL,
  `metaGroupName` varchar(100) DEFAULT NULL,
  `description` varchar(1000) DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  PRIMARY KEY (`metaGroupID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invMetaGroups`
--

LOCK TABLES `invMetaGroups` WRITE;
/*!40000 ALTER TABLE `invMetaGroups` DISABLE KEYS */;
/*!40000 ALTER TABLE `invMetaGroups` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `invMetaTypes`
--

DROP TABLE IF EXISTS `invMetaTypes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `invMetaTypes` (
  `typeID` int(11) NOT NULL,
  `parentTypeID` int(11) DEFAULT NULL,
  `metaGroupID` int(11) DEFAULT NULL,
  PRIMARY KEY (`typeID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invMetaTypes`
--

LOCK TABLES `invMetaTypes` WRITE;
/*!40000 ALTER TABLE `invMetaTypes` DISABLE KEYS */;
/*!40000 ALTER TABLE `invMetaTypes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `invNames`
--

DROP TABLE IF EXISTS `invNames`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `invNames` (
  `itemID` int(11) NOT NULL,
  `itemName` varchar(200) NOT NULL,
  PRIMARY KEY (`itemID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invNames`
--

LOCK TABLES `invNames` WRITE;
/*!40000 ALTER TABLE `invNames` DISABLE KEYS */;
/*!40000 ALTER TABLE `invNames` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `invPositions`
--
-- 

-- 
--
-- Table structure for table `invTraits`
--
-- 

-- 
--
-- Table structure for table `invTypeMaterials`
--

DROP TABLE IF EXISTS `invTypeMaterials`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `invTypeMaterials` (
  `typeID` int(11) NOT NULL,
  `materialTypeID` int(11) NOT NULL,
  `quantity` int(11) NOT NULL,
  PRIMARY KEY (`typeID`,`materialTypeID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invTypeMaterials`
--

LOCK TABLES `invTypeMaterials` WRITE;
/*!40000 ALTER TABLE `invTypeMaterials` DISABLE KEYS */;
/*!40000 ALTER TABLE `invTypeMaterials` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `invTypeReactions`
--
-- 

-- 
--
-- Table structure for table `invTypes`
--

DROP TABLE IF EXISTS `invTypes`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `invTypes` (
  `typeID` int(11) NOT NULL,
  `groupID` int(11) DEFAULT NULL,
  `typeName` varchar(100) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `mass` double DEFAULT NULL,
  `volume` double DEFAULT NULL,
  `capacity` double DEFAULT NULL,
  `portionSize` int(11) DEFAULT NULL,
  `raceID` int(11) DEFAULT NULL,
  `basePrice` decimal(19,4) DEFAULT NULL,
  `published` tinyint(1) DEFAULT NULL,
  `marketGroupID` int(11) DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  `soundID` int(11) DEFAULT NULL,
  `graphicID` int(11) DEFAULT NULL,
  PRIMARY KEY (`typeID`),
  KEY `ix_invTypes_groupID` (`groupID`),
  CONSTRAINT `invtype_published` CHECK (`published` in (0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invTypes`
--

LOCK TABLES `invTypes` WRITE;
/*!40000 ALTER TABLE `invTypes` DISABLE KEYS */;
/*!40000 ALTER TABLE `invTypes` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `invUniqueNames`
--

DROP TABLE IF EXISTS `invUniqueNames`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `invUniqueNames` (
  `itemID` int(11) NOT NULL,
  `itemName` varchar(200) NOT NULL,
  `groupID` int(11) DEFAULT NULL,
  PRIMARY KEY (`itemID`),
  UNIQUE KEY `ix_invUniqueNames_itemName` (`itemName`),
  KEY `invUniqueNames_IX_GroupName` (`groupID`,`itemName`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `invUniqueNames`
--

LOCK TABLES `invUniqueNames` WRITE;
/*!40000 ALTER TABLE `invUniqueNames` DISABLE KEYS */;
/*!40000 ALTER TABLE `invUniqueNames` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `invVolumes`
--
-- 

-- 
--
-- Table structure for table `mapCelestialGraphics`
--
-- 

-- 
--
-- Table structure for table `mapCelestialStatistics`
--
-- 

-- 
--
-- Table structure for table `mapConstellationJumps`
--
-- 

-- 
--
-- Table structure for table `mapConstellations`
--

DROP TABLE IF EXISTS `mapConstellations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `mapConstellations` (
  `regionID` int(11) DEFAULT NULL,
  `constellationID` int(11) NOT NULL,
  `constellationName` varchar(100) DEFAULT NULL,
  `x` double DEFAULT NULL,
  `y` double DEFAULT NULL,
  `z` double DEFAULT NULL,
  `xMin` double DEFAULT NULL,
  `xMax` double DEFAULT NULL,
  `yMin` double DEFAULT NULL,
  `yMax` double DEFAULT NULL,
  `zMin` double DEFAULT NULL,
  `zMax` double DEFAULT NULL,
  `factionID` int(11) DEFAULT NULL,
  `radius` float DEFAULT NULL,
  PRIMARY KEY (`constellationID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `mapConstellations`
--

LOCK TABLES `mapConstellations` WRITE;
/*!40000 ALTER TABLE `mapConstellations` DISABLE KEYS */;
/*!40000 ALTER TABLE `mapConstellations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `mapDenormalize`
--

DROP TABLE IF EXISTS `mapDenormalize`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `mapDenormalize` (
  `itemID` int(11) NOT NULL,
  `typeID` int(11) DEFAULT NULL,
  `groupID` int(11) DEFAULT NULL,
  `solarSystemID` int(11) DEFAULT NULL,
  `constellationID` int(11) DEFAULT NULL,
  `regionID` int(11) DEFAULT NULL,
  `orbitID` int(11) DEFAULT NULL,
  `x` double DEFAULT NULL,
  `y` double DEFAULT NULL,
  `z` double DEFAULT NULL,
  `radius` double DEFAULT NULL,
  `itemName` varchar(100) DEFAULT NULL,
  `security` double DEFAULT NULL,
  `celestialIndex` int(11) DEFAULT NULL,
  `orbitIndex` int(11) DEFAULT NULL,
  PRIMARY KEY (`itemID`),
  KEY `ix_mapDenormalize_solarSystemID` (`solarSystemID`),
  KEY `ix_mapDenormalize_orbitID` (`orbitID`),
  KEY `mapDenormalize_IX_groupConstellation` (`groupID`,`constellationID`),
  KEY `mapDenormalize_IX_groupRegion` (`groupID`,`regionID`),
  KEY `ix_mapDenormalize_regionID` (`regionID`),
  KEY `mapDenormalize_IX_groupSystem` (`groupID`,`solarSystemID`),
  KEY `ix_mapDenormalize_typeID` (`typeID`),
  KEY `ix_mapDenormalize_constellationID` (`constellationID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `mapDenormalize`
--

LOCK TABLES `mapDenormalize` WRITE;
/*!40000 ALTER TABLE `mapDenormalize` DISABLE KEYS */;
/*!40000 ALTER TABLE `mapDenormalize` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `mapJumps`
--
-- 

-- 
--
-- Table structure for table `mapLandmarks`
--

DROP TABLE IF EXISTS `mapLandmarks`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `mapLandmarks` (
  `landmarkID` int(11) NOT NULL,
  `landmarkName` varchar(100) DEFAULT NULL,
  `description` text DEFAULT NULL,
  `locationID` int(11) DEFAULT NULL,
  `x` double DEFAULT NULL,
  `y` double DEFAULT NULL,
  `z` double DEFAULT NULL,
  `iconID` int(11) DEFAULT NULL,
  PRIMARY KEY (`landmarkID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `mapLandmarks`
--

LOCK TABLES `mapLandmarks` WRITE;
/*!40000 ALTER TABLE `mapLandmarks` DISABLE KEYS */;
/*!40000 ALTER TABLE `mapLandmarks` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `mapLocationScenes`
--
-- 

-- 
--
-- Table structure for table `mapLocationWormholeClasses`
--
-- 

-- 
--
-- Table structure for table `mapRegionJumps`
--
-- 

-- 
--
-- Table structure for table `mapRegions`
--

DROP TABLE IF EXISTS `mapRegions`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `mapRegions` (
  `regionID` int(11) NOT NULL,
  `regionName` varchar(100) DEFAULT NULL,
  `x` double DEFAULT NULL,
  `y` double DEFAULT NULL,
  `z` double DEFAULT NULL,
  `xMin` double DEFAULT NULL,
  `xMax` double DEFAULT NULL,
  `yMin` double DEFAULT NULL,
  `yMax` double DEFAULT NULL,
  `zMin` double DEFAULT NULL,
  `zMax` double DEFAULT NULL,
  `factionID` int(11) DEFAULT NULL,
  `nebula` int(11) DEFAULT NULL,
  `radius` float DEFAULT NULL,
  PRIMARY KEY (`regionID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `mapRegions`
--

LOCK TABLES `mapRegions` WRITE;
/*!40000 ALTER TABLE `mapRegions` DISABLE KEYS */;
/*!40000 ALTER TABLE `mapRegions` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `mapSolarSystemJumps`
--
-- 

-- 
--
-- Table structure for table `mapSolarSystems`
--

DROP TABLE IF EXISTS `mapSolarSystems`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `mapSolarSystems` (
  `regionID` int(11) DEFAULT NULL,
  `constellationID` int(11) DEFAULT NULL,
  `solarSystemID` int(11) NOT NULL,
  `solarSystemName` varchar(100) DEFAULT NULL,
  `x` double DEFAULT NULL,
  `y` double DEFAULT NULL,
  `z` double DEFAULT NULL,
  `xMin` double DEFAULT NULL,
  `xMax` double DEFAULT NULL,
  `yMin` double DEFAULT NULL,
  `yMax` double DEFAULT NULL,
  `zMin` double DEFAULT NULL,
  `zMax` double DEFAULT NULL,
  `luminosity` double DEFAULT NULL,
  `border` tinyint(1) DEFAULT NULL,
  `fringe` tinyint(1) DEFAULT NULL,
  `corridor` tinyint(1) DEFAULT NULL,
  `hub` tinyint(1) DEFAULT NULL,
  `international` tinyint(1) DEFAULT NULL,
  `regional` tinyint(1) DEFAULT NULL,
  `constellation` tinyint(1) DEFAULT NULL,
  `security` double DEFAULT NULL,
  `factionID` int(11) DEFAULT NULL,
  `radius` double DEFAULT NULL,
  `sunTypeID` int(11) DEFAULT NULL,
  `securityClass` varchar(2) DEFAULT NULL,
  PRIMARY KEY (`solarSystemID`),
  KEY `ix_mapSolarSystems_regionID` (`regionID`),
  KEY `ix_mapSolarSystems_security` (`security`),
  KEY `ix_mapSolarSystems_constellationID` (`constellationID`),
  CONSTRAINT `mapss_border` CHECK (`border` in (0,1)),
  CONSTRAINT `mapss_fringe` CHECK (`fringe` in (0,1)),
  CONSTRAINT `mapss_corridor` CHECK (`corridor` in (0,1)),
  CONSTRAINT `mapss_hub` CHECK (`hub` in (0,1)),
  CONSTRAINT `mapss_internat` CHECK (`international` in (0,1)),
  CONSTRAINT `mapss_regional` CHECK (`regional` in (0,1)),
  CONSTRAINT `mapss_constel` CHECK (`constellation` in (0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `mapSolarSystems`
--

LOCK TABLES `mapSolarSystems` WRITE;
/*!40000 ALTER TABLE `mapSolarSystems` DISABLE KEYS */;
/*!40000 ALTER TABLE `mapSolarSystems` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `mapUniverse`
--
-- 

-- 
--
-- Table structure for table `planetSchematics`
--

DROP TABLE IF EXISTS `planetSchematics`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `planetSchematics` (
  `schematicID` int(11) NOT NULL,
  `schematicName` varchar(255) DEFAULT NULL,
  `cycleTime` int(11) DEFAULT NULL,
  PRIMARY KEY (`schematicID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `planetSchematics`
--

LOCK TABLES `planetSchematics` WRITE;
/*!40000 ALTER TABLE `planetSchematics` DISABLE KEYS */;
/*!40000 ALTER TABLE `planetSchematics` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `planetSchematicsPinMap`
--

DROP TABLE IF EXISTS `planetSchematicsPinMap`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `planetSchematicsPinMap` (
  `schematicID` int(11) NOT NULL,
  `pinTypeID` int(11) NOT NULL,
  PRIMARY KEY (`schematicID`,`pinTypeID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `planetSchematicsPinMap`
--

LOCK TABLES `planetSchematicsPinMap` WRITE;
/*!40000 ALTER TABLE `planetSchematicsPinMap` DISABLE KEYS */;
/*!40000 ALTER TABLE `planetSchematicsPinMap` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `planetSchematicsTypeMap`
--

DROP TABLE IF EXISTS `planetSchematicsTypeMap`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `planetSchematicsTypeMap` (
  `schematicID` int(11) NOT NULL,
  `typeID` int(11) NOT NULL,
  `quantity` int(11) DEFAULT NULL,
  `isInput` tinyint(1) DEFAULT NULL,
  PRIMARY KEY (`schematicID`,`typeID`),
  CONSTRAINT `pstm_input` CHECK (`isInput` in (0,1))
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `planetSchematicsTypeMap`
--

LOCK TABLES `planetSchematicsTypeMap` WRITE;
/*!40000 ALTER TABLE `planetSchematicsTypeMap` DISABLE KEYS */;
/*!40000 ALTER TABLE `planetSchematicsTypeMap` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ramActivities`
--
-- 

-- 
--
-- Table structure for table `ramAssemblyLineStations`
--
-- 

-- 
--
-- Table structure for table `ramAssemblyLineTypeDetailPerCategory`
--
-- 

-- 
--
-- Table structure for table `ramAssemblyLineTypeDetailPerGroup`
--
-- 

-- 
--
-- Table structure for table `ramAssemblyLineTypes`
--
-- 

-- 
--
-- Table structure for table `ramInstallationTypeContents`
--
-- 

-- 
--
-- Table structure for table `skinLicense`
--

DROP TABLE IF EXISTS `skinLicense`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `skinLicense` (
  `licenseTypeID` int(11) NOT NULL,
  `duration` int(11) DEFAULT NULL,
  `skinID` int(11) DEFAULT NULL,
  PRIMARY KEY (`licenseTypeID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `skinLicense`
--

LOCK TABLES `skinLicense` WRITE;
/*!40000 ALTER TABLE `skinLicense` DISABLE KEYS */;
/*!40000 ALTER TABLE `skinLicense` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `skinMaterials`
--

DROP TABLE IF EXISTS `skinMaterials`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `skinMaterials` (
  `skinMaterialID` int(11) NOT NULL,
  `displayNameID` int(11) DEFAULT NULL,
  `materialSetID` int(11) DEFAULT NULL,
  PRIMARY KEY (`skinMaterialID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `skinMaterials`
--

LOCK TABLES `skinMaterials` WRITE;
/*!40000 ALTER TABLE `skinMaterials` DISABLE KEYS */;
/*!40000 ALTER TABLE `skinMaterials` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `skinShip`
--

DROP TABLE IF EXISTS `skinShip`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `skinShip` (
  `skinID` int(11) DEFAULT NULL,
  `typeID` int(11) DEFAULT NULL,
  KEY `ix_skinShip_skinID` (`skinID`),
  KEY `ix_skinShip_typeID` (`typeID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `skinShip`
--

LOCK TABLES `skinShip` WRITE;
/*!40000 ALTER TABLE `skinShip` DISABLE KEYS */;
/*!40000 ALTER TABLE `skinShip` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `skins`
--
-- 

-- 
--
-- Table structure for table `staOperationServices`
--
-- 

-- 
--
-- Table structure for table `staOperations`
--
-- 

-- 
--
-- Table structure for table `staServices`
--
-- 

-- 
--
-- Table structure for table `staStationTypes`
--
-- 

-- 
--
-- Table structure for table `staStations`
--

DROP TABLE IF EXISTS `staStations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `staStations` (
  `stationID` bigint(20) NOT NULL,
  `security` double DEFAULT NULL,
  `dockingCostPerVolume` double DEFAULT NULL,
  `maxShipVolumeDockable` double DEFAULT NULL,
  `officeRentalCost` int(11) DEFAULT NULL,
  `operationID` int(11) DEFAULT NULL,
  `stationTypeID` int(11) DEFAULT NULL,
  `corporationID` int(11) DEFAULT NULL,
  `solarSystemID` int(11) DEFAULT NULL,
  `constellationID` int(11) DEFAULT NULL,
  `regionID` int(11) DEFAULT NULL,
  `stationName` varchar(100) DEFAULT NULL,
  `x` double DEFAULT NULL,
  `y` double DEFAULT NULL,
  `z` double DEFAULT NULL,
  `reprocessingEfficiency` double DEFAULT NULL,
  `reprocessingStationsTake` double DEFAULT NULL,
  `reprocessingHangarFlag` int(11) DEFAULT NULL,
  PRIMARY KEY (`stationID`),
  KEY `ix_staStations_stationTypeID` (`stationTypeID`),
  KEY `ix_staStations_constellationID` (`constellationID`),
  KEY `ix_staStations_corporationID` (`corporationID`),
  KEY `ix_staStations_regionID` (`regionID`),
  KEY `ix_staStations_solarSystemID` (`solarSystemID`),
  KEY `ix_staStations_operationID` (`operationID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `staStations`
--

LOCK TABLES `staStations` WRITE;
/*!40000 ALTER TABLE `staStations` DISABLE KEYS */;
/*!40000 ALTER TABLE `staStations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `translationTables`
--
-- 

-- 
--
-- Table structure for table `trnTranslationColumns`
--
-- 

-- 
--
-- Table structure for table `trnTranslationLanguages`
--
-- 

-- 
--
-- Table structure for table `trnTranslations`
--

DROP TABLE IF EXISTS `trnTranslations`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `trnTranslations` (
  `tcID` int(11) NOT NULL,
  `keyID` int(11) NOT NULL,
  `languageID` varchar(50) NOT NULL,
  `text` text NOT NULL,
  PRIMARY KEY (`tcID`,`keyID`,`languageID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `trnTranslations`
--

LOCK TABLES `trnTranslations` WRITE;
/*!40000 ALTER TABLE `trnTranslations` DISABLE KEYS */;
/*!40000 ALTER TABLE `trnTranslations` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `warCombatZoneSystems`
--
-- 

-- 
--
-- Table structure for table `warCombatZones`
--
-- 

-- /*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;
-- 
-- /*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
-- /*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
-- /*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
-- /*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
-- /*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
-- /*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
-- /*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;
-- 
-- Dump completed on 2025-07-07 17:47:07
-- 