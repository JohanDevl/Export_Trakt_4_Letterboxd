# Stratégie de Gestion des Images Docker

## Vue d'ensemble

Le système de gestion des images Docker a été optimisé pour maintenir uniquement les images nécessaires et nettoyer automatiquement les images obsolètes.

## Stratégie de Tagging

### 🏷️ Main Branch
- `latest` - Toujours la dernière version stable
- `main` - Tag fixe pour la branche main
- `v1.2.3` - Version sémantique exacte (générée automatiquement)

### 🏷️ Develop Branch  
- `develop` - Toujours la dernière version de développement

### 🏷️ Pull Requests
- `PR-123` - Image pour tester une PR spécifique avant merge

## Registres Supportés

- **Docker Hub**: `johandevl/export-trakt-4-letterboxd`
- **GitHub Container Registry**: `ghcr.io/johandevl/export_trakt_4_letterboxd`

## Système de Nettoyage Automatique

### 🧹 Nettoyage PR (déclenché à la fermeture de PR)
- Supprime automatiquement l'image `PR-{numero}` des deux registres
- Se déclenche seulement quand une PR sur `main` ou `develop` est fermée

### 🧹 Nettoyage Programmé (quotidien à 2h UTC)
- Nettoie les images obsolètes automatiquement
- Préserve les tags protégés :
  - `latest`, `main`, `develop`
  - Toutes les versions sémantiques (`v1.2.3`)
  - Les tags des PR ouvertes (`PR-123`)
- Supprime tout le reste

## Utilisation

### Pour tester une PR :
```bash
docker pull johandevl/export-trakt-4-letterboxd:PR-123
```

### Pour utiliser la dernière version stable :
```bash
docker pull johandevl/export-trakt-4-letterboxd:latest
```

### Pour utiliser la version de développement :
```bash
docker pull johandevl/export-trakt-4-letterboxd:develop
```

### Pour utiliser une version spécifique :
```bash
docker pull johandevl/export-trakt-4-letterboxd:v1.2.3
```

## Workflows GitHub Actions

- **docker-build.yml** : Construit et publie les images
- **docker-cleanup.yml** : Nettoie les images obsolètes
- **auto-tag.yml** : Crée automatiquement les versions sémantiques

## Avantages

✅ **Images PR disponibles** pour tests pré-merge  
✅ **Nettoyage automatique** des images obsolètes  
✅ **Double registre** (Docker Hub + GitHub Container Registry)  
✅ **Versioning sémantique** automatique  
✅ **Conservation intelligente** des versions importantes  
✅ **Nettoyage quotidien** programmé  