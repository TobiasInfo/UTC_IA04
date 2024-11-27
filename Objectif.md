Simulation d'une journée sur un festival 
- phase d'arrivée 
- phase concert
- phase déplacement 
- phase fin

Les festivaliers se déplacent selon la phase de la journée, avec à chaque fois un objectif donné à chacun. Ils ne se déplacent pas de manière totalement aléatoire. 
Un taux de malaise est fixé pour tous les festivaliers.
Une durée de survie en cas de malaise est donnée à tous les festivaliers (unique ou aléatoire). Si une personne ne reçoit pas de trousse de soin à temps, on considère qu'elle aura des séquelles et c'est donc un échec. 
Pour représenter les risques plus élevés d'accidents, coup de chaleur, malaise, crise cardiaque dans les situations stressantes et dans les foules, le taux de malaise est également influencé par le nombre de personnes autour d'un festivalier.


Drones qui communiquent entre eux en P2P
Leur objectif est de sauver tout le monde
On va mesurer : 
- temps de sauvetage des victimes 
- nombre de personnes non pris en charge avant la fin de leur durée de survie

On va utiliser la simulation pour montrer 3 protocoles des drones, et les performances selon des paramètres à changer. La simulation est un moyen de mesurer la pertinence de chaque protocole.
Les paramètres à faire varier sont :
- configuration du festival (obstacles, lieu de recharge, lieu de récupération trousse à pharmacie)
- nombre de festivaliers
- taux de malaise (forte chaleur ou non)
Les drones peuvent aussi varier : (pas sûr qu'on veuille faire ça)
- autonomie
- vitesse
- protocole 

Premier protocole pour les drones :

- des positions initiales leurs sont assignées par un opérateur 
- chaque drone patrouille sa zone
- lorsque une personne est identifié comme "en danger" le drone part chercher une trousse à pharmacie et lui rapporte
- lorsque la batterie tombe sous les XX% il retourne échanger sa batterie à la base. 

Point à surveiller dans ce cas : 
- que se passe-t-il s'il voit une victime mais n'a pas l'autonomie pour aller chercher la trousse de secours. 
- que se passe-t-il lorsque le drone va changer de batterie mais voit une victime sur la route ? Ignore ? 
- etc


Choses qui peuvent être modifiées pour avoir des protocoles différents :

- les drones partent de la base, et scannent la zone. Ils se répartissent de manière autonome des zones propres à chacun. 

- Un drone reste en stand by pour remplacer un drone qui irait se recharger, ou le reste du temps apporter la trousse de secours aux victimes détectées par un autre drone

- lorsqu'un drone part chercher une batterie ou la trousse de secours, les drones avoisinant se partagent sa zone à surveiller de manière temporaire 

- on rajoute une fonctionnalité de zone d'importance, pour que certains drones aient des zones plus petites mais avec plus de personnes, et inversement 
-
