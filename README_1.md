# Système Multi-Drones pour la Sécurité d'Événements Festifs

## Table des Matières
1. [Introduction](#introduction)
2. [Architecture du Projet](#architecture-du-projet)
3. [Environnement et Interactions](#environnement-et-interactions)
4. [Modélisation des Agents](#modélisation-des-agents)
5. [Interface Graphique de Simulation](#interface-graphique-de-simulation)
6. [Analyse par Lots et Résultats](#analyse-par-lots-et-résultats)

## Introduction

Les festivals de grande envergure présentent des défis majeurs en termes de sécurité et de gestion des urgences médicales. Notre système propose une solution basée sur une flotte de drones autonomes collaborant avec des équipes de secours au sol pour assurer une surveillance continue et une intervention rapide.

Le système repose sur trois types d'agents :
- Les drones de surveillance, équipés de systèmes de détection et de communication
- Les équipes de secours, intervenant sur le terrain
- Les festivaliers, avec leurs comportements et besoins

## Architecture du Projet

```text
UTC_IA04/
├── cmd/
│   ├── run_simulations/          # Exécution des simulations benchmark
│   │   ├── results/              # Stockage des résultats d'analyse
│   │   └── main.go              # Point d'entrée benchmark
│   ├── simu/                    # Simulation graphique
│   │   ├── drawutils.go         # Utilitaires de dessin
│   │   └── simu.go             # Logique de simulation
│   └── ui/                      # Interface utilisateur
│       ├── assets/              # Ressources graphiques
│       ├── components/          # Composants réutilisables
│       ├── constants/           # Constantes UI
│       ├── button.go           # Gestion des boutons
│       ├── liste_deroulante.go # Menus déroulants
│       ├── textfield.go        # Champs de texte
│       └── main_gui_ebiten.go  # Point d'entrée interface graphique
├── configs/                     # Configurations des cartes
├── pkg/                        # Logique métier
│   ├── entities/               # Agents autonomes
│   ├── models/                 # Structures de données
│   └── simulation/             # Moteur de simulation
└── vendor/                     # Dépendances externes
```

## Environnement et Interactions

### Le Terrain du Festival

L'environnement de simulation reproduit la configuration d'un festival avec trois zones distinctes :

La zone d'entrée constitue le point d'accès des festivaliers. Elle permet de contrôler le flux d'entrée des participants et d'établir le premier niveau de surveillance.

La zone principale concentre l'essentiel des activités et des points d'intérêt :
- Scènes de spectacle
- Stands de restauration et de boissons
- Zones de repos
- Installations sanitaires
- Postes de secours
- Stations de recharge pour les drones

La zone de sortie permet une gestion ordonnée des départs.

### Dynamique Temporelle

La simulation utilise un ratio temporel de 1:60, où une seconde réelle correspond à une minute simulée. Cette compression permet d'observer l'évolution d'un festival complet tout en maintenant une précision suffisante pour l'analyse des interventions.

## Modélisation des Agents

### Les Festivaliers

Chaque festivalier possède un profil qui influence son comportement :

1. L'Aventurier
- Grande mobilité dans l'espace
- Exploration active des différentes zones
- Niveau de fatigue augmentant rapidement

2. Le Prudent
- Préfère les zones moins denses
- Maintient une distance de sécurité importante
- Progression méthodique entre les points d'intérêt

3. Le Social
- Tendance à suivre les groupes
- Préférence pour les zones animées
- Interactions fréquentes avec les points d'intérêt

4. L'Indépendant
- Parcours personnalisé du site
- Faible influence des mouvements de foule
- Rythme d'activité régulier

Le système modélise la fatigue et les risques de malaise selon :
```python
P(malaise) = P_base x (1 - Resistance_Malaise) x (1 - Niveau_Energie)
où P_base = 0.005
```
### Les Drones de Surveillance

Les drones constituent le cœur du système de détection. Chaque drone est un agent autonome disposant des capacités suivantes :

1. Capacités de Base
- Un système de détection avec une portée configurable (DroneSeeRange)
- Un système de communication avec une portée définie (DroneCommRange)
- Une gestion autonome de l'énergie avec :
  - Surveillance du niveau de batterie
  - Recherche de points de recharge
  - Planification des recharges

2. Détection et Surveillance
Le drone effectue une surveillance continue de sa zone assignée. La probabilité de détection d'une personne en détresse suit la formule :
```go
probaDetection := max(0, 1.0/float64(s.DroneSeeRange)-(float64(nbPersDetected)*0.03))
```
Cette formule modélise la diminution de l'efficacité de détection avec la distance et le nombre de personnes déjà détectées.

3. Patrouille et Communication
Le drone maintient une patrouille systématique de sa zone. En cas de détection d'une personne en détresse, il peut :
- Alerter directement un point de secours si à portée
- Relayer l'information via d'autres drones
- Coordonner une intervention avec les équipes au sol
  
### Les Équipes de Secours

Les sauveteurs représentent l'interface entre la surveillance automatisée et l'intervention humaine. Positionnés dans des postes de secours stratégiques, ils :
- Reçoivent les alertes des drones
- Se déplacent vers les personnes en détresse
- Administrent les premiers soins
- Retournent à leur poste après intervention

### Protocoles de Communication des Drones

#### Protocole 1 : Système de Base

Le protocole 1 implémente les mécanismes fondamentaux du système. Il définit les capacités individuelles des drones :

##### Fonctionnalités Implémentées
- Scan continu de la zone de surveillance du drone
- Détection des personnes en détresse
- Mémorisation des cas détectés dans une liste interne
- Déplacement vers le point de secours le plus proche en cas de détection
- Gestion autonome de la batterie avec recherche de point de recharge quand nécessaire

#### Protocole 2 : Communication Locale

Le protocole 2 ajoute au protocole 1 les fonctionnalités suivantes :

##### Nouvelles Fonctionnalités
- Implémentation d'un pattern de patrouille en zigzag remplaçant le mouvement aléatoire
- Établissement de communication entre drones à portée directe
- Capacité de transmission des informations aux drones voisins
- Fonction de transfert de responsabilité entre drones proches
- Mécanisme de délégation des cas détectés aux drones mieux positionnés

##### Mécanismes Techniques Ajoutés
- Vérification de la portée de communication entre drones
- Système de transfert de données entre drones à portée
- Algorithme de patrouille structurée
- Protocole de délégation des responsabilités

#### Protocole 3 : Réseau Multi-Sauts

Le protocole 3 étend le protocole 2 avec les fonctionnalités réseau suivantes :

##### Extensions Techniques
- Implémentation d'un réseau de communication maillé entre drones
- Communication possible au-delà de la portée directe via des relais
- Formation dynamique de sous-réseaux de communication
- Transmission d'informations à travers le réseau de drones
- Coordination via le réseau pour atteindre les points de secours

##### Structures de Données Ajoutées
- Tables de routage pour la communication multi-sauts
- Base de données distribuée des cas détectés
- Graphe des connexions entre drones
- Système de propagation des messages à travers le réseau

#### Protocole 4 : Optimisation du Réseau

Le protocole 4 complète le protocole 3 avec ces mécanismes d'optimisation :

##### Fonctionnalités Additionnelles
- Calcul des distances effectives aux points de secours pour chaque drone
- Sélection automatique du drone le plus proche pour chaque intervention
- Distribution optimisée des responsabilités dans le réseau
- Transfert intelligent des cas selon la topologie du réseau
- Prise en compte de la distance au point de secours dans les décisions


## Interface Graphique de Simulation

### Configuration Initiale
Pour lancer la simulation :
1. Cloner le projet :
```bash
git clone https://github.com/TobiasInfo/UTC_IA04
cd UTC_IA04/cmd
go run ./main_gui_ebiten.go
```

### Écran d'Accueil
L'interface permet de configurer :
- Le nombre de drones détermine la capacité de surveillance du système. Un équilibre doit être trouvé entre une couverture suffisante et une utilisation efficiente des ressources.

- La population initiale de festivaliers influence directement la complexité des interactions et la charge sur le système de surveillance.

- La sélection de la carte définit la disposition physique du festival, avec ses zones et points d'intérêt spécifiques.

- Le choix du protocole de communication des drones impacte significativement leur efficacité collective.

### Vue Principale
L'interface graphique, développée avec le moteur Ebiten, offre une visualisation claire et interactive de la simulation. Elle se compose de plusieurs éléments clés :

La vue principale présente une représentation en temps réel du festival. Les festivaliers, les drones et les points d'intérêt sont représentés par des icônes distinctives. Les drones affichent leur champ de vision sous forme d'un cercle d'ombre, permettant de visualiser la couverture de surveillance.

Le panneau de contrôle permet de :
- Mettre en pause la simulation
- Avancer pas à pas en mode debug
- Visualiser les métriques en temps réel

Deux visualisations dynamiques enrichissent l'analyse :

La carte de densité (à gauche) représente la distribution des festivaliers sur le site. Cette visualisation peut être agrandie pour une analyse plus détaillée des mouvements de foule.

Le graphe de réseau (à droite) illustre les communications entre drones et leur connexion avec les points de secours. Il permet de comprendre la topologie du réseau et d'identifier d'éventuelles zones de faible couverture.


## Analyse par Lots et Résultats

Cette section présente l'analyse exhaustive du système via des simulations non graphiques.


## Configuration des Tests

Le système effectue une analyse exhaustive en testant toutes les combinaisons possibles des paramètres suivants :

### Paramètres Variables
- **Nombre de drones** : 2, 5, et 10 drones
  - 2 drones représente une couverture minimale
  - 5 drones offre une couverture moyenne
  - 10 drones permet une couverture intensive
  
- **Nombre de festivaliers** : 200, 500, et 1000 personnes
  - 200 personnes simule un petit événement
  - 500 personnes représente un événement moyen
  - 1000 personnes teste le système en charge élevée

- **Protocoles** : 1, 2, 3, et 4
  - Protocole 1 : système baseline avec communication simple
  - Protocole 2 : ajout de la patrouille structurée
  - Protocole 3 : introduction du réseau de communication
  - Protocole 4 : optimisation du réseau et des décisions

- **Configurations de carte** : 
  - festival_layout_1 : configuration avec point de secours sur le côté
  - festival_layout_2 : configuration avec deux points de secours
  - festival_layout_3 : configuration avec point de secours central

Au total, cela représente 108 configurations différentes (3 x 3 x 4 x 3 = 108 configurations), chacune exécutée 5 fois pour assurer la fiabilité statistique.

## Structure des Résultats

Le programme génère un dossier `results` organisé comme suit :

```text
results/
├── 2d_200p_p1_festival_layout_1/    # Configuration minimale, protocole 1, carte 1
├── 2d_200p_p1_festival_layout_2/
├── ...
├── 5d_500p_p2_festival_layout_1/    # Configuration moyenne, protocole 2
├── 5d_500p_p2_festival_layout_2/
├── ...
└── 10d_1000p_p4_festival_layout_3/  # Configuration maximale, protocole 4, carte 3
```

Chaque dossier de configuration contient :
```text
configuration_folder/
├── run_1_metrics.txt            # Données de la première exécution
├── run_2_metrics.txt
├── run_3_metrics.txt
├── run_4_metrics.txt
├── run_5_metrics.txt
├── metrics.txt                  # Moyennes et analyses statistiques
├── rescue_stats_people.png      # Graphique temporel des sauvetages
└── rescue_stats_time.png        # Graphique des temps de réponse
```

## Métriques Analysées

### Par Exécution (run_X_metrics.txt)
Chaque fichier d'exécution enregistre :
```text
Run X Results
================
Total People: [nombre]
People in Distress: [nombre]
Cases Treated: [nombre]
Cases Dead: [nombre]
Average Battery: [pourcentage]%
Average Coverage: [pourcentage]%
Runtime: [durée]
Total Ticks: [nombre]
```

### Analyse Globale (metrics.txt)
Le fichier de synthèse comprend :
```text
Simulation Results (Averaged over 5 runs)
=====================================
Total People: [moyenne]
People in Distress: [moyenne]
Cases Treated: [moyenne]
Cases Dead: [moyenne]
Average Battery: [moyenne]%
Average Coverage: [moyenne]%
Average Runtime: [durée moyenne]
Total Ticks: [moyenne]

Performance Metrics:
- Treatment Success Rate: [pourcentage]%
- Mortality Rate: [pourcentage]%
- Average Response Time: [durée]
```
## Visualisations Générées

### Graphique de Sauvetages (rescue_stats_people.png)
Ce graphique présente deux courbes principales :
- En rouge : l'évolution du nombre de personnes en détresse
- En vert : l'évolution du nombre de personnes sauvées
L'axe des abscisses représente le temps de simulation en ticks, permettant d'observer les moments critiques et l'efficacité des interventions.

### Graphique des Temps de Réponse (rescue_stats_time.png)
Ce graphique montre une courbe bleue représentant l'évolution du temps moyen de sauvetage au cours de la simulation. Il permet d'évaluer si le système maintient son efficacité même sous charge.

Cette analyse complète permet d'optimiser :
- Le dimensionnement de la flotte de drones
- Le choix du protocole selon le contexte
- Le positionnement des points de secours
- L'allocation des ressources d'intervention
