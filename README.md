# 🎪 Système Multi-Drones pour la Sécurité d'Événements Festifs

## 📑 Table des Matières
1. [Introduction](#-introduction)
2. [Introduction](#-démarrage-et-lancement-du-projet)
3. [Architecture du Projet](#-architecture-du-projet)
4. [Environnement et Interactions](#-environnement-et-interactions)
5. [Implémentation](#-implémentation-)
6. [Modélisation des Agents](#-modélisation-des-agents)
7. [Interface Graphique de Simulation](#-interface-graphique-de-simulation)
8. [Analyse par Lots et Résultats](#-analyse-par-lots-et-résultats)
9. [Bibliographie](#-bibliographie)

## 🚀 Introduction

### Vue d'ensemble

Les festivals de grande envergure présentent des défis majeurs en termes de sécurité et de gestion des urgences médicales. Notre système propose une solution basée sur une flotte de drones autonomes collaborant avec des équipes de secours au sol pour assurer une surveillance continue et une intervention rapide.

#### Principaux Agents
- 🛸 Les drones de surveillance, équipés de systèmes de détection et de communication
- 🚑 Les équipes de secours, intervenant sur le terrain
- 👥 Les festivaliers, avec leurs comportements et besoins


## 🏗 Architecture du Projet

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

## 🏁 Démarrage et Lancement du Projet

### 🛠️ Prérequis

- **Go** (version 1.23.3+)
- **Git**

### 📥 Installation

Il faut éxecuter les commandes suivantes pour installer le projet :

   ```bash
   git clone https://github.com/TobiasInfo/UTC_IA04.git
   cd UTC_IA04
   go mod vendor
   go mod tidy
   ```

### 🚀 Lancer la Simulation

- Pour lancer la simulation avec l'interface Graphique :

  ```bash
  go run ./cmd/ui/main_gui_ebiten.go
  ```

### 🔧 Configuration Supplémentaire

- Modifier les configurations des maps dans `configs/`
- Ajuster les paramètres via le GUI du démarrage.

### 🧹 Nettoyage

- Nettoyer les modules

  ```bash
  go mod tidy
  ```

## 🌍 Environnement et Interactions

### 🎯 Le Terrain du Festival

L'environnement de simulation reproduit la configuration d'un festival avec trois zones distinctes :

#### Zone d'Entrée
La zone d'entrée constitue le point d'accès des festivaliers. Elle permet de contrôler le flux d'entrée des participants et d'établir le premier niveau de surveillance.

#### Zone Principale
La zone principale concentre l'essentiel des activités et des points d'intérêt :
- 🎭 Scènes de spectacle
- 🍕 Stands de restauration et de boissons
- 🛋 Zones de repos
- 🚽 Installations sanitaires
- 🏥 Postes de secours
- 🔋 Stations de recharge pour les drones

#### Zone de Sortie
La zone de sortie permet une gestion ordonnée des départs.

### ⏱ Dynamique Temporelle

La simulation utilise un ratio temporel de 1:60, où une seconde réelle correspond à une minute simulée. Cette compression permet d'observer l'évolution d'un festival complet tout en maintenant une précision suffisante pour l'analyse des interventions.

## 💻 Implémentation 

Les Agents utilisent une boucle de Perception/Délibération/Action, et évoluent en parallèle avec des goroutines pour permettre une évolution indépendante et non-déterministe dans la mesure des fonctionnalités du langage go.  

Il a été choisi de synchroniser les agents pour ne leur permettre qu'une itération de leur cycle de perception/délibération/action par tick de la simulation globale pour conserver une cohérence des actions des agents entre eux, et rester plus fidèle aux conditions réelles.

Un objet Simulation contient l'ensemble des éléments utiles à notre simulation, dont une instance de Carte, qui mémorise et gère les positions et déplacements des agents.

Pour l'interface graphique l'outil Ebiten a été utilisé, pour permettre une implémentation globale 100% en Go.

Les images utilisées ont été générées par des IA génératives, puis retouchées ensuite à la main.

## 🤖 Modélisation des Agents

### 👥 Les Festivaliers

Chaque festivalier possède un profil qui influence son comportement :

#### 1. L'Aventurier 🏃
- Grande mobilité dans l'espace
- Exploration active des différentes zones
- Niveau de fatigue augmentant rapidement

#### 2. Le Prudent 🚶
- Préfère les zones moins denses
- Maintient une distance de sécurité importante
- Progression méthodique entre les points d'intérêt

#### 3. Le Social 👯
- Tendance à suivre les groupes
- Préférence pour les zones animées
- Interactions fréquentes avec les points d'intérêt

#### 4. L'Indépendant 🧘
- Parcours personnalisé du site
- Faible influence des mouvements de foule
- Rythme d'activité régulier

Ces profils ont une influence sur :  
- La vitesse de déplacement de l'individu  
- les variables "CrowdFollowingTendency" et "PersonalSpace" qui influe sur la tendance de l'individu à aller ou non dans les zones avec de nombreux autres individus  
- Le niveau d'énergie à partir duquel un individu passera en mode repos  
- La résistance au malaise de l'individu  
- L'intérêt porté par l'individu à chaque POI, et donc vers lesquels il préférera se diriger.

Lorsque qu'un participant atteint un POI, il va y rester pendant une durée variable, puis repartir à la recherche d'un autre POI.

Le système modélise la fatigue et les risques de malaise selon :
```python
P(malaise) = P_base x (1 - Resistance_Malaise) x (1 - Niveau_Energie)
où P_base = 0.005
```
Lorsqu'un participant en situation de détresse a été remarqué, l'information doit être remontée à la tente infirmerie qui détachera ensuite un pompier qui ira sauver le participant.

### 🛸 Les Drones de Surveillance

Les drones constituent le cœur du système de détection. Chaque drone est un agent autonome disposant des capacités suivantes :

#### 1. Capacités de Base
- Un système de détection avec une portée configurable (DroneSeeRange)
- Un système de communication avec une portée définie (DroneCommRange)
- Une gestion autonome de l'énergie avec :
  - Surveillance du niveau de batterie
  - Recherche de points de recharge
  - Planification des recharges

#### 2. 🎯 Détection et Surveillance
Le drone effectue une surveillance continue de sa zone assignée. La probabilité de détection d'une personne en détresse suit la formule :
```go
probaDetection := max(0, 1.0/float64(s.DroneSeeRange)-(float64(nbPersDetected)*0.03))
```
Cette formule modélise la diminution de l'efficacité de détection avec la distance et le nombre de personnes déjà détectées.

#### 3. 📡 Patrouille et Communication
Le drone maintient une patrouille systématique de sa zone. En cas de détection d'une personne en détresse, il peut :
- Alerter directement un point de secours si à portée
- Relayer l'information via d'autres drones
- Coordonner une intervention avec les équipes au sol

### 🚑 Les Équipes de Secours

Les sauveteurs représentent l'interface entre la surveillance automatisée et l'intervention humaine. Positionnés dans des postes de secours stratégiques, ils :
- Reçoivent les alertes des drones
- Se déplacent vers les personnes en détresse
- Administrent les premiers soins
- Retournent à leur poste après intervention

### 📡 Protocoles de Communication des Drones

#### 🔰 Protocole 1 : Système de Base

Le protocole 1 implémente les mécanismes fondamentaux du système. Il définit les capacités individuelles des drones :

##### Fonctionnalités Implémentées
- Scan continu de la zone de surveillance du drone
- Détection des personnes en détresse
- Mémorisation des cas détectés dans une liste interne
- Déplacement vers le point de secours le plus proche en cas de détection
- Gestion autonome de la batterie avec recherche de point de recharge quand nécessaire

#### 🔄 Protocole 2 : Communication Locale

Le protocole 2 ajoute au protocole 1 les fonctionnalités suivantes :

##### Nouvelles Fonctionnalités
- Implémentation d’un schéma de patrouille en zigzag, remplaçant les mouvements aléatoires, pour assurer un quadrillage efficace de toute la zone.
- Mise en place de communications entre drones dans un rayon de portée directe.
- Capacité de transmettre des informations aux drones voisins pour maintenir la fluidité des échanges.
- Fonctionnalité de transfert de responsabilité entre drones à proximité immédiate.
- Mécanisme de délégation des cas détectés aux drones mieux positionnés, y compris ceux situés au niveau supérieur hiérarchique (n+1).

##### Mécanismes Techniques Ajoutés
- Vérification de la portée de communication entre drones
- Système de transfert de données entre drones à portée
- Algorithme de patrouille structurée
- Mécanisme de délégation des responsabilités

#### 🌐 Protocole 3 : Réseau Multi-Sauts

Le protocole 3 étend le protocole 2 avec les fonctionnalités réseau suivantes :

##### Extensions Techniques
- Implémentation d’un réseau de communication maillé entre drones (chaque drone appartient à un sous-réseau, équivalent à un sous-graphe de drones).
- Communication possible au-delà de la portée directe grâce à un système de relais (les drones se relaient les informations entre eux).
- Formation dynamique de sous-réseaux de communication (les sous-réseaux correspondent à des sous-graphes).
- Transmission d’informations à travers le réseau de drones.
- Si aucun drone ne peut transmettre l’information directement à un point de relais principal (RP), le drone ayant détecté l’incident prend en charge la mission de se déplacer pour informer le RP.

#### ⚡ Protocole 4 : Optimisation du Réseau

Le protocole 4 complète le protocole 3 avec ces mécanismes d'optimisation :

##### Fonctionnalités Additionnelles
- Calcul des distances effectives aux points de secours pour chaque drone
- Sélection automatique du drone le plus proche pour chaque intervention
- Distribution optimisée des responsabilités dans le réseau
- Transfert intelligent des cas selon la topologie du réseau
- Prise en compte de la distance au point de secours dans les décisions

## 🎮 Interface Graphique de Simulation

### ⚙️ Configuration Initiale
Pour lancer la simulation :
```bash
git clone https://github.com/TobiasInfo/UTC_IA04
cd UTC_IA04/cmd
go run ./main_gui_ebiten.go
```

### 🏁 Écran d'Accueil
L'interface permet de configurer :
- Le nombre de drones détermine la capacité de surveillance du système. Un équilibre doit être trouvé entre une couverture suffisante et une utilisation efficiente des ressources.

- La population initiale de festivaliers influence directement la complexité des interactions et la charge sur le système de surveillance.

- La sélection de la carte définit la disposition physique du festival, avec ses zones et points d'intérêt spécifiques.

- Le choix du protocole de communication des drones impacte significativement leur efficacité collective.

### 🖥️ Vue Principale
L'interface graphique, développée avec le moteur Ebiten, offre une visualisation claire et interactive de la simulation. Elle se compose de plusieurs éléments clés :

La vue principale présente une représentation en temps réel du festival. Les festivaliers, les drones et les points d'intérêt sont représentés par des icônes distinctives. Les drones affichent leur champ de vision sous forme d'un cercle d'ombre, permettant de visualiser la couverture de surveillance.

Le panneau de contrôle permet de :
- ⏸️ Mettre en pause la simulation
- 🔍 Avancer pas à pas en mode debug
- 📊 Visualiser les métriques en temps réel

Deux visualisations dynamiques enrichissent l'analyse :

- La carte de densité (à gauche) représente la distribution des festivaliers sur le site. Cette visualisation peut être agrandie pour une analyse plus détaillée des mouvements de foule.
- Le graphe de réseau (à droite) illustre les communications entre drones et leur connexion avec les points de secours. Il permet de comprendre la topologie du réseau et d'identifier d'éventuelles zones de faible couverture.

Pour évaluer les performances de la flotte de drone, une fois la simulation terminée deux graphiques sont également générés et sauvegardés:

- Le premier graphique représente l'évolution du nombre de personnes en situation de détresse, ainsi que les moments de prise en charge des personnes en fonction du temps.  
- Le second graphique représente pour chaque personne sauvée, le temps pris pour le sauvetage. On a ainsi une estimation du temps nécessaire entre le début d'un malaise et l'arrivée d'un secouriste auprès du participant, pour chaque protocole de drone.

## 📊 Analyse par Lots et Résultats

Cette section présente l'outil d'analyse par lots (benchmarking) développé pour évaluer systématiquement les performances du système multi-drones sans interface graphique. Contrairement à la simulation visuelle qui permet une observation qualitative, cet outil fournit une analyse quantitative approfondie des différentes configurations.

### 🔍 Vue d'ensemble

L'analyse par lots s'exécute via le fichier `main.go` et automatise l'exécution de multiples simulations avec différentes combinaisons de paramètres. Pour chaque configuration, l'outil :
1. Lance 5 simulations identiques
2. Collecte les métriques détaillées
3. Calcule les moyennes et écarts
4. Génère des visualisations des résultats
5. Exporte les données dans une structure organisée

Pour lancer l'analyse :
```bash
cd UTC_IA04
go run main.go
```

### 🎛️ Paramètres d'Analyse

L'outil teste systématiquement les combinaisons des paramètres suivants :

#### Taille de la Flotte de Drones
- **2 drones** : Couverture minimale pour tester la résilience
- **5 drones** : Configuration moyenne, équilibre coût/efficacité
- **10 drones** : Couverture intensive pour événements majeurs

#### Population de Festivaliers
- **200 personnes** : Petits événements, charge faible
- **500 personnes** : Événements moyens, charge normale
- **1000 personnes** : Grands événements, charge élevée

#### Protocoles de Communication
- **Protocole 1** : Système de base, communication directe
- **Protocole 2** : Patrouille structurée et communication locale
- **Protocole 3** : Communication multi-sauts en réseau
- **Protocole 4** : Optimisation du réseau et des décisions

#### Configurations de Carte
- **festival_layout_1** : Point de secours latéral
- **festival_layout_2** : Double points de secours
- **festival_layout_3** : Point de secours central

Au total, l'analyse couvre 108 configurations uniques (3×3×4×3), chacune répétée 5 fois pour assurer la significativité statistique.

### 📂 Structure des Résultats

L'outil génère une hiérarchie de dossiers dans `./results/` organisée comme suit :

```text
results/
├── {n}d_{p}p_p{x}_{layout}/    # Un dossier par configuration
│   ├── metrics.txt             # Synthèse statistique
│   ├── rescue_stats_people.png # Évolution des sauvetages
│   ├── rescue_stats_time.png   # Temps de réponse
│   ├── run_1_metrics.txt       # Détails par simulation
│   ├── run_2_metrics.txt
│   ├── run_3_metrics.txt
│   ├── run_4_metrics.txt
│   └── run_5_metrics.txt
```

Où :
- `n` : nombre de drones (2, 5, 10)
- `p` : population (200, 500, 1000)
- `x` : numéro de protocole (1-4)
- `layout` : configuration de carte

### 📊 Métriques Analysées

#### Métriques Globales (metrics.txt)
```text
Simulation Results (Averaged over 5 runs)
=====================================
Total People: [moyenne]
People in Distress: [moyenne]
Cases Treated: [moyenne]
Cases Dead: [moyenne]
Average Battery: [moyenne]%
Average Coverage: [moyenne]%
Average Runtime: [durée]
Total Ticks: [ticks]

Performance Metrics:
- Treatment Success Rate: [pourcentage]%
- Mortality Rate: [pourcentage]%
- Average Response Time: [durée]
```

#### Métriques Détaillées (run_X_metrics.txt)
Chaque simulation individuelle génère un rapport détaillé incluant :
- Statistiques complètes de population
- États des drones (batterie, couverture)
- Temps de réponse aux incidents
- Durée totale de simulation

### 📊 Visualisations Générées

#### Évolution des Sauvetages (rescue_stats_people.png)
Graphique temporel montrant :
- **Courbe rouge** : Nombre de personnes en détresse
- **Courbe verte** : Nombre de personnes sauvées
Permet d'identifier les pics d'activité et l'efficacité des interventions.

#### Analyse des Temps de Réponse (rescue_stats_time.png)
- **Courbe bleue** : Temps moyen de sauvetage
- Permet d'évaluer la réactivité du système et sa stabilité sous charge

### Utilisation des Résultats

Ces analyses permettent de :
1. Optimiser le dimensionnement de la flotte
2. Sélectionner le protocole le plus adapté selon le contexte
3. Valider le positionnement des points de secours
4. Identifier les configurations critiques
5. Estimer les ressources nécessaires selon la taille de l'événement

Les résultats fournissent une base quantitative pour les décisions de déploiement et l'amélioration continue du système.

## 📚 Bibliographie

### Systèmes Multi-Agents
- Ferber, J. (1999). *Multi-Agent Systems: An Introduction to Distributed Artificial Intelligence*. Addison-Wesley.
- Wooldridge, M. (2009). *An Introduction to MultiAgent Systems*. John Wiley & Sons.
- Weiss, G. (2013). *Multiagent Systems*. MIT Press.

### Drones et Surveillance
- Floreano, D., & Wood, R. J. (2015). *Science, technology and the future of small autonomous drones*. Nature, 521(7553), 460-466.
- Shakhatreh, H., et al. (2019). *Unmanned Aerial Vehicles (UAVs): A Survey on Civil Applications and Key Research Challenges*. IEEE Access.

### Gestion d'Événements et Sécurité
- Still, G. K. (2014). *Introduction to Crowd Science*. CRC Press.
- Fruin, J. J. (1993). *The causes and prevention of crowd disasters*. Engineering for Crowd Safety.

### Technologies et Implémentation
- Kennedy, Alan A. (2019). *Go in Practice*. Manning Publications.
- Butcher, Matt. (2017). *Go in Action*. Manning Publications.
- Documentation Ebiten: [https://ebiten.org/documents.html](https://ebiten.org/documents.html)

### Algorithmes et Optimisation
- Cormen, T. H., et al. (2009). *Introduction to Algorithms*. MIT Press.
- Bonabeau, E., et al. (1999). *Swarm Intelligence: From Natural to Artificial Systems*. Oxford University Press.
