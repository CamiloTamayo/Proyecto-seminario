-- MariaDB dump 10.19-11.3.0-MariaDB, for Win64 (AMD64)
--
-- Host: localhost    Database: seminariodb
-- ------------------------------------------------------
-- Server version	11.3.0-MariaDB

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `maquina_fisica`
--

DROP TABLE IF EXISTS `maquina_fisica`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `maquina_fisica` (
  `idmf` int(11) NOT NULL AUTO_INCREMENT,
  `bridge_adapter` varchar(100) DEFAULT NULL,
  `cpu` int(11) DEFAULT NULL,
  `hostname` varchar(50) DEFAULT NULL,
  `ip` varchar(50) DEFAULT NULL,
  `os` varchar(50) DEFAULT NULL,
  `rammb` int(11) DEFAULT NULL,
  `storagegb` int(11) DEFAULT NULL,
  PRIMARY KEY (`idmf`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `maquina_fisica`
--

LOCK TABLES `maquina_fisica` WRITE;
/*!40000 ALTER TABLE `maquina_fisica` DISABLE KEYS */;
INSERT INTO `maquina_fisica` VALUES
(1,'Qualcomm Atheros QCA9377 Wireless Network Adapter',6,'Luz Stella','192.168.1.68','Windows 10 Home',4098,2048);
/*!40000 ALTER TABLE `maquina_fisica` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `maquina_virtual`
--

DROP TABLE IF EXISTS `maquina_virtual`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `maquina_virtual` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `contrasenia` varchar(500) DEFAULT NULL,
  `estado` varchar(50) DEFAULT NULL,
  `hostname` varchar(50) DEFAULT NULL,
  `ip` varchar(50) DEFAULT NULL,
  `nombre` varchar(50) DEFAULT NULL,
  `mfisica_idmf` int(11) DEFAULT NULL,
  `tipo_maquina_id` int(11) DEFAULT NULL,
  `usuario_id` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `FKb8b2p2krjmyfqe5ssubadkewj` (`mfisica_idmf`),
  KEY `FK2cbdbuxp42exv4eu3jw5olyh2` (`tipo_maquina_id`),
  KEY `FKr3vc6iw935emt8krouxwvc1ym` (`usuario_id`),
  CONSTRAINT `FK2cbdbuxp42exv4eu3jw5olyh2` FOREIGN KEY (`tipo_maquina_id`) REFERENCES `tipo_maquina` (`id`),
  CONSTRAINT `FKb8b2p2krjmyfqe5ssubadkewj` FOREIGN KEY (`mfisica_idmf`) REFERENCES `maquina_fisica` (`idmf`),
  CONSTRAINT `FKr3vc6iw935emt8krouxwvc1ym` FOREIGN KEY (`usuario_id`) REFERENCES `usuario` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `maquina_virtual`
--

LOCK TABLES `maquina_virtual` WRITE;
/*!40000 ALTER TABLE `maquina_virtual` DISABLE KEYS */;
/*!40000 ALTER TABLE `maquina_virtual` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `tipo_maquina`
--

DROP TABLE IF EXISTS `tipo_maquina`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tipo_maquina` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `cpu` int(11) DEFAULT NULL,
  `hostname` varchar(50) DEFAULT NULL,
  `nombre` varchar(50) DEFAULT NULL,
  `rammb` int(11) DEFAULT NULL,
  `sistema_operativo` varchar(50) DEFAULT NULL,
  `storagegb` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `tipo_maquina`
--

LOCK TABLES `tipo_maquina` WRITE;
/*!40000 ALTER TABLE `tipo_maquina` DISABLE KEYS */;
INSERT INTO `tipo_maquina` VALUES
(1,2,'vmtipo1','vmtipo1',1024,'Debian 11 server',20),
(2,4,'vmtipo1','vmtipo1',2048,'Debian 11 server',30);
/*!40000 ALTER TABLE `tipo_maquina` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `tipo_usuario`
--

DROP TABLE IF EXISTS `tipo_usuario`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `tipo_usuario` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `nombre` varchar(50) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `tipo_usuario`
--

LOCK TABLES `tipo_usuario` WRITE;
/*!40000 ALTER TABLE `tipo_usuario` DISABLE KEYS */;
INSERT INTO `tipo_usuario` VALUES
(1,'administrador'),
(2,'estudiante'),
(3,'docente'),
(4,'unlogged');
/*!40000 ALTER TABLE `tipo_usuario` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `usuario`
--

DROP TABLE IF EXISTS `usuario`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8 */;
CREATE TABLE `usuario` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `apellidos` varchar(255) DEFAULT NULL,
  `contrasenia` varchar(255) DEFAULT NULL,
  `correo` varchar(255) DEFAULT NULL,
  `nombre` varchar(255) DEFAULT NULL,
  `tipo_usuario_id` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `FKe581tp719p3d7o5u2w9sre10b` (`tipo_usuario_id`),
  CONSTRAINT `FKe581tp719p3d7o5u2w9sre10b` FOREIGN KEY (`tipo_usuario_id`) REFERENCES `tipo_usuario` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `usuario`
--

LOCK TABLES `usuario` WRITE;
/*!40000 ALTER TABLE `usuario` DISABLE KEYS */;
INSERT INTO `usuario` VALUES
(1,'Tamayo Amariles','$2a$10$ED2jBd.R7BJl/NSlCHP8pO9fyWyNZ36QOYAWm2j2xBambr1Cc3XlS','juanc.tamayoa@uqvirtual.edu.co','Juan Camilo',1);
/*!40000 ALTER TABLE `usuario` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2023-11-23 12:03:03
