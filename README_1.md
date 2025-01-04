# Système Multi-Drones pour la Sécurité d'Événements Festifs

## Introduction : Un Festival Sous Haute Surveillance

Les festivals de grande envergure représentent des défis majeurs en termes de sécurité et de gestion des urgences médicales. Dans ces environnements dynamiques où des milliers de personnes se rassemblent, la rapidité d'intervention en cas de malaise ou d'incident est cruciale. Notre système propose une approche novatrice : une flotte de drones autonomes travaillant en synergie avec des équipes de secours au sol pour assurer une surveillance continue et une intervention rapide.

Dans cet écosystème complexe, trois types d'agents interagissent constamment : les festivaliers, avec leurs comportements et besoins variés, les drones de surveillance, équipés de systèmes de détection sophistiqués, et les équipes de secours, prêtes à intervenir sur le terrain. Cette simulation modélise leurs interactions dans un environnement festival réaliste, permettant d'étudier et d'optimiser les stratégies d'intervention.

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
## Les Acteurs du Système

### Les Festivaliers : Des Comportements Humains Complexes

Les festivaliers constituent le cœur vivant de notre simulation. Chaque participant est modélisé comme un agent autonome doté d'une personnalité unique qui influence ses décisions et ses actions. Nous avons identifié quatre profils psychologiques distincts :

L'Aventurier (Adventurous) se caractérise par une grande mobilité et une tendance à explorer l'ensemble du site. Moins sensible à la densité de la foule, il présente néanmoins un risque accru de fatigue dû à son activité intense.

Le Prudent (Cautious) privilégie les zones calmes et maintient une distance confortable avec les autres participants. Son comportement méthodique réduit les risques de malaise, mais peut limiter son expérience du festival.

Le Social recherche activement les zones animées et les rassemblements. Sa tendance à suivre la foule influence significativement ses déplacements, créant des dynamiques de groupe intéressantes à observer.

L'Indépendant se démarque par son autonomie dans ses choix de déplacement et d'activités. Moins influencé par les mouvements de foule, il suit son propre parcours sur le site.

Chaque festivalier possède également un niveau d'énergie qui évolue au fil du temps et des activités. Le système modélise la fatigue et les risques de malaise selon la formule :

```python
P(malaise) = P_base x (1 - Resistance_Malaise) x (1 - Niveau_Energie)
où P_base = 0.005
```

### Les Protocoles de Communication des Drones

#### Protocole 1 : Système de Base

Le protocole 1 implémente les mécanismes fondamentaux du système. Il définit les capacités individuelles des drones :

##### Fonctionnalités Implémentées
- Scan continu de la zone de surveillance du drone
- Détection des personnes en détresse selon la formule :
```go
probaDetection := max(0, 1.0/float64(s.DroneSeeRange)-(float64(nbPersDetected)*0.03))
```
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

### Les Équipes de Secours : L'Interface Humaine

Les sauveteurs représentent le lien crucial entre la surveillance automatisée et l'intervention humaine. Basés dans des postes de secours stratégiquement positionnés, ils réagissent aux alertes transmises par les drones pour porter assistance aux festivaliers en détresse.

## Environnement et Interactions

### Le Terrain du Festival

L'environnement de simulation reproduit fidèlement la configuration d'un festival avec trois zones distinctes :

La zone d'entrée constitue le point d'accès principal des festivaliers. Elle joue un rôle crucial dans la gestion des flux de participants.

La zone principale concentre l'essentiel des activités et des points d'intérêt. Elle est parsemée de différents POI (Points of Interest) :
- Scènes de spectacle
- Stands de restauration et de boissons
- Zones de repos
- Installations sanitaires
- Postes de secours
- Stations de recharge pour les drones

La zone de sortie permet aux participants de quitter le site de manière fluide et contrôlée.

### Dynamique Temporelle

Le temps dans la simulation s'écoule de manière accélérée, avec un ratio de 1:60 (une seconde réelle équivaut à une minute simulée). Cette compression temporelle permet d'observer l'évolution complète d'un festival tout en maintenant une granularité suffisante pour l'analyse des interventions d'urgence.

## Utilisation du Système

### Lancement et Configuration

L'interface de simulation offre un contrôle précis sur les paramètres de l'expérience. Pour démarrer, l'utilisateur doit :

1. Télécharger le projet depuis le dépôt GitHub (https://github.com/TobiasInfo/UTC_IA04)
2. Naviguer vers le répertoire d'exécution : `cd XXX\UTC_IA04\cmd`
3. Lancer l'application : `go run .\main_gui_ebiten.go`

L'écran d'accueil permet de configurer les paramètres essentiels de la simulation :

Le nombre de drones détermine la capacité de surveillance du système. Un équilibre doit être trouvé entre une couverture suffisante et une utilisation efficiente des ressources.

La population initiale de festivaliers influence directement la complexité des interactions et la charge sur le système de surveillance.

La sélection de la carte définit la disposition physique du festival, avec ses zones et points d'intérêt spécifiques.

Le choix du protocole de communication des drones impacte significativement leur efficacité collective.

### Interface de Simulation

L'interface graphique, développée avec le moteur Ebiten, offre une visualisation claire et interactive de la simulation. Elle se compose de plusieurs éléments clés :

La vue principale présente une représentation en temps réel du festival. Les festivaliers, les drones et les points d'intérêt sont représentés par des icônes distinctives. Les drones affichent leur champ de vision sous forme d'un cercle d'ombre, permettant de visualiser la couverture de surveillance.

Le panneau de contrôle permet de :
- Mettre en pause la simulation
- Avancer pas à pas en mode debug
- Visualiser les métriques en temps réel

Deux visualisations dynamiques enrichissent l'analyse :

La carte de densité (à gauche) représente la distribution des festivaliers sur le site. Cette visualisation peut être agrandie pour une analyse plus détaillée des mouvements de foule.

Le graphe de réseau (à droite) illustre les communications entre drones et leur connexion avec les points de secours. Il permet de comprendre la topologie du réseau et d'identifier d'éventuelles zones de faible couverture.

### Métriques et Analyses

Le système collecte et analyse en temps réel de nombreuses métriques :

Les métriques de population incluent :
- Le nombre total de participants
- Les cas de détresse actifs
- Les interventions réussies
- Les cas non traités à temps

Les métriques opérationnelles comprennent :
- Le niveau moyen de batterie des drones
- Le pourcentage de couverture du terrain
- L'efficacité des communications
- Les temps de réponse aux incidents

À la fin de chaque simulation, le système génère des graphiques d'analyse détaillés :
- Évolution temporelle des cas de détresse
- Temps de réponse pour chaque intervention
- Efficacité des protocoles de coordination
- Couverture spatiale des interventions

## Sorties et Données Générées

Chaque session de simulation produit :
- Statistiques complètes d'intervention
- Métriques de performance réseau
- Analyses temporelles des incidents
- Cartes de chaleur de couverture
- Graphiques comparatifs des protocoles

Les données sont automatiquement sauvegardées dans le répertoire du projet à la fin de chaque simulation pour permettre une analyse ultérieure détaillée.
