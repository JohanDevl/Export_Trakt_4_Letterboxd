# CLAUDE.md - Main CLI Application

## Module Overview

Ce module contient le point d'entrée principal de l'application Export Trakt 4 Letterboxd. Il gère l'interface en ligne de commande (CLI), l'orchestration des différents modules, et l'exécution des commandes utilisateur.

## Responsabilités

### 🚀 Point d'Entrée Principal
- **Initialisation de l'application** : Configuration des logs, traducteur, gestionnaires de sécurité
- **Gestion des flags CLI** : Parsing et validation des arguments de ligne de commande
- **Orchestration des modules** : Coordination entre API, auth, export, scheduler, etc.
- **Gestion des erreurs** : Recovery des paniques et gestion centralisée des erreurs

### 🔧 Commandes Disponibles

#### Commandes d'Export
- **`export`** : Export standard des données (watched, collection, shows, ratings, watchlist, all)
- **`--run`** : Exécution unique d'export avec sortie immédiate
- **`--export`** : Type d'export (watched, collection, shows, ratings, watchlist, all)
- **`--mode`** : Mode d'export (normal, initial, complete)
- **`--history-mode`** : Mode historique (aggregated, individual) pour les films regardés

#### Commandes d'Authentification
- **`auth`** : Authentification OAuth 2.0 interactive avec serveur de callback local
- **`token-status`** : Vérification du statut des tokens OAuth
- **`token-refresh`** : Rafraîchissement manuel des tokens
- **`token-clear`** : Suppression des tokens stockés
- **`auth-url`** : Génération d'URL d'authentification pour usage manuel
- **`auth-code`** : Authentification avec code d'autorisation fourni

#### Commandes de Serveur
- **`server`** : Serveur HTTP persistant avec interface web et callbacks OAuth
- **`--schedule`** : Planification cron pour exports automatiques

#### Commandes Utilitaires
- **`validate`** : Validation de la configuration
- **`--validate-security`** : Validation complète des configurations de sécurité
- **`fix-permissions`** : Correction des permissions des fichiers credentials
- **`setup`** : Instructions de configuration initiale

### 📊 Modes d'Export

#### Mode Agrégé (par défaut)
```bash
./export_trakt --run --export watched --history-mode aggregated
```
- Une entrée par film unique avec la date de visionnage la plus récente
- Compatible avec le comportement original Letterboxd
- Optimisé pour la vitesse et la simplicité

#### Mode Historique Individuel
```bash
./export_trakt --run --export watched --history-mode individual
```
- Une entrée par événement de visionnage
- Historique complet de tous les visionnages
- Suivi précis des re-visionnages avec chronologie

### 🔒 Gestion de la Sécurité

#### Initialisation Sécurisée
- **Security Manager** : Validation et application des politiques de sécurité
- **Keyring Manager** : Gestion des credentials avec différents backends (system, env, file, memory)
- **Token Manager** : Gestion sécurisée des tokens OAuth avec chiffrement

#### Validation de Sécurité
```bash
./export_trakt --validate-security
```
Vérifie :
- Configuration de sécurité
- Fonctionnement du chiffrement/déchiffrement
- Sanitisation des entrées
- Protection contre le path traversal
- Permissions des fichiers
- Configuration HTTPS
- Audit logging
- Rate limiting

### 🕒 Système de Planification

#### Planification Cron
```bash
./export_trakt --schedule "0 */6 * * *" --export all --mode complete
```
- Support des expressions cron standard
- Gestion des fuseaux horaires configurables
- Logs détaillés des exécutions
- Arrêt gracieux avec SIGINT/SIGTERM

#### Serveur Persistant avec Planification
```bash
./export_trakt server --schedule "0 */6 * * *" --export all
```
- Interface web complète avec pagination
- Exports automatiques en arrière-plan
- Callbacks OAuth intégrés

### 🌐 Interface Web

#### Fonctionnalités
- **Dashboard** : Vue d'ensemble des exports et statut d'authentification
- **Page d'Exports** : Liste paginée des exports avec téléchargements
- **Statut** : Monitoring en temps réel des services
- **Authentification** : Flux OAuth complet via interface web

#### Endpoints Principaux
- `GET /` : Dashboard principal
- `GET /exports` : Page d'exports avec pagination
- `GET /status` : Statut des services
- `GET /auth-url` : Génération d'URL d'authentification
- `GET /download/{file}` : Téléchargement des exports

### 🔄 Gestion des Tokens OAuth

#### Flux d'Authentification Complet
1. **Génération d'URL** : Création d'URL d'autorisation avec PKCE
2. **Serveur de Callback** : Serveur local temporaire pour recevoir le code
3. **Échange de Token** : Conversion code → access/refresh tokens
4. **Stockage Sécurisé** : Sauvegarde dans keyring avec chiffrement
5. **Rafraîchissement Automatique** : Renouvellement transparent des tokens

#### Backends de Stockage
- **system** : Keyring système (macOS Keychain, Windows Credential Manager, Linux Secret Service)
- **env** : Variables d'environnement
- **file** : Fichier chiffré avec AES-256
- **memory** : Stockage en mémoire (non persistant)

### 📈 Performance et Resilience

#### Optimisations
- **Worker Pool** : Traitement concurrent des requêtes API
- **LRU Cache** : Cache des réponses API pour réduire les appels
- **Streaming** : Traitement en mémoire optimisé pour gros datasets
- **Circuit Breaker** : Protection contre les pannes en cascade

#### Monitoring
- **Logs Structurés** : Avec clés de traduction i18n
- **Métriques** : Prometheus et OpenTelemetry
- **Health Checks** : Vérifications de santé des services

### 🌍 Internationalisation

Support multilingue avec traducteur intégré :
- **Anglais** (par défaut)
- **Français**
- **Allemand** 
- **Espagnol**

### 💾 Gestion des Exports

#### Scanning des Exports Existants
- **Détection Automatique** : Scan des dossiers d'export timestampés
- **Métadonnées** : Extraction du type, taille, nombre d'enregistrements
- **Historique** : Tri chronologique des exports

#### Types d'Export Supportés
- **watched** : Films regardés (mode agrégé ou individuel)
- **collection** : Collection personnelle
- **shows** : Séries TV avec épisodes
- **ratings** : Notes attribuées
- **watchlist** : Liste de films à voir
- **all** : Export complet de tous les types

## Patterns d'Architecture

### Injection de Dépendances
```go
// Initialisation ordonnée des composants
tokenManager := auth.NewTokenManager(cfg, log, keyringMgr)
traktClient := api.NewClientWithTokenManager(cfg, log, tokenManager)
letterboxdExporter := export.NewLetterboxdExporter(cfg, log)
```

### Gestion d'Erreurs Structurée
```go
// Utilisation des clés de traduction pour les logs
log.Error("errors.api_request_failed", map[string]interface{}{"error": err.Error()})
```

### Configuration Centralisée
```go
// Priorisation : flags CLI > env vars > config.toml > defaults
cfg, err := config.LoadConfig(*configPath)
```

## Commandes de Développement

### Tests et Validation
```bash
# Validation de sécurité
./export_trakt --validate-security

# Test d'export unique
./export_trakt --run --export watched --mode complete

# Serveur de développement
./export_trakt server --schedule "*/2 * * * *" --export watched
```

### Debug et Monitoring
```bash
# Logs verbeux
EXPORT_VERBOSE=true ./export_trakt --schedule "*/1 * * * *"

# Niveau de log personnalisé
LOG_LEVEL=debug ./export_trakt server
```

## Notes de Sécurité

- **Récupération de Panique** : Protection contre les crashes non gérés
- **Validation d'Entrée** : Sanitisation de tous les inputs utilisateur
- **Gestion des Permissions** : Vérification et correction automatique des permissions fichiers
- **HTTPS Enforcement** : Application optionnelle du HTTPS pour toutes les communications
- **Audit Trail** : Logging détaillé des actions sensibles

## Points d'Extension

Pour ajouter de nouvelles commandes :
1. Ajouter le parsing du flag dans `main()`
2. Ajouter le case dans le switch de commandes
3. Implémenter la fonction handler correspondante
4. Ajouter les clés de traduction i18n si nécessaire

Le module sert de point d'orchestration central et peut facilement être étendu pour de nouvelles fonctionnalités tout en maintenant la cohérence architecturale.