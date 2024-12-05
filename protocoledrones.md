# Protocoles de Surveillance par Drones pour Festival

## Caractéristiques Communes des Drones
- Communication P2P basique :
  * Partage des positions des drones
  * Partage des incidents détectés
  * Partage du niveau de batterie
  * Partage du statut actuel (en patrouille/en charge/en livraison)

- Contraintes de ressources :
  * Batterie limitée
  * Nécessité de collecter le matériel médical au poste médical
  * Nécessité de recharger aux stations de charge

## 1. Protocole "Emergency First"
### Gestion de la batterie
- Retour à la station de charge si batterie ≤ 20%
- Reprise de la patrouille si batterie ≥ 80%

### Flux d'intervention
1. Détection d'incident
2. Collection du matériel au poste médical le plus proche
3. Livraison du matériel à la personne en détresse
4. Retour à la patrouille

### Priorités de déplacement
1. Retour à la charge si batterie critique
2. Collecte du matériel médical si incident détecté
3. Livraison du matériel à l'incident
4. Patrouille aléatoire si aucun incident

## 2. Protocole Collaboratif
### Drones de Patrouille (60% de la flotte)
- Gestion batterie :
  * Retour charge à 25%
  * Reprise patrouille à 85%
- Missions :
  * Surveillance de zone
  * Signalement des incidents aux drones de secours
  * Intervention uniquement si aucun drone de secours disponible

### Drones de Secours (40% de la flotte)
- Gestion batterie :
  * Retour charge à 30%
  * Reprise service à 90%
- Flux d'intervention :
  1. Réception alerte incident
  2. Collection matériel médical
  3. Livraison à l'incident
  4. Retour position centrale

### Différences Principales
- Emergency First : Tous les drones font tout
- Collaboratif : Rôles spécialisés avec seuils de batterie différents