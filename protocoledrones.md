# Explication des protocoles de drones

## Fonctionnement du protocole 1 :

Step 1 :  
- Je scanne les personnes en danger  
- Si je vois une personne en danger, je la sauvegarde.  

Step 2 :
- Dès que ma liste est supérieur > 1 je m'en vais vers le RP le + Proche pour régler les problémes.  
- Si je n'ai plus de batterie, je bouge vers le point de charge le plus proche.  
    - J'essaye lors de mon mouvement de transmettre ma liste à mes voisins pour qu'ils aillent informer le rescurer à ma place.  
- Une fois que ma charge est terminée, je bouge vers le point de sauvetage le plus proche.  


## Fonctionnement du protocole 2 :

Step 0 :  
- Si je n'ai plus de batterie, je bouge vers le point de charge le plus proche.  
    - J'essaye lors de mon mouvement de transmettre ma liste à mes voisins pour qu'ils aillent informer le rescurer à ma place.  
- Une fois que ma charge est terminée, je bouge vers le point de sauvetage le plus proche.  

Step 1 :  
- Je scanne les personnes en danger  
- Si je vois une personne en danger, je la sauvegarde.  

Step 2 :  
- J'essaye de communiquer avec un RP si un RP est dans mon rayon de communication.  
   - Si aucun RP n'est dans mon rayon de communication.  
		- J'essaye de voir si je peux envoyer l'information à un drone qui est en n+1 de mon rayon de communication.  
		- Si je ne peux pas, je bouge vers le rescue point le plus proche.  
- Je bouge vers le rescue point si je ne peux pas communiquer.  


## Fonctionnement du protocole 3 :  

Step 0 :  
- Si je n'ai plus de batterie, je bouge vers le point de charge le plus proche.  
    - J'essaye lors de mon mouvement de transmettre ma liste à mes voisins pour qu'ils aillent informer le rescurer à ma place.  
- Une fois que ma charge est terminée, je bouge vers le point de sauvetage le plus proche.  

Step 1 :  
- Je scanne les personnes en danger  
- Si je vois une personne en danger, je la sauvegarde.  

Step 2 :  
- J'essaye de communiquer avec un RP si un RP est dans mon rayon de communication.  
   - Si aucun RP n'est dans mon rayon de communication.  
		- J'essaye de voir si je peux envoyer l'information à un drone qui est dans mon network.  
			- Un network est un sous-ensemble de drones qui peuvent communiquer entre eux, ils sont chainées et ils forment un sous-graphe.  
		- Si je ne peux pas, je bouge vers le rescue point le plus proche.  
- Je bouge vers le rescue point si je ne peux pas communiquer.  

## Fonctionnement du protocole 4 :

Step 0 :  
- Si je n'ai plus de batterie, je bouge vers le point de charge le plus proche.  
    - J'essaye lors de mon mouvement de transmettre ma liste à mes voisins pour qu'ils aillent informer le rescurer à ma place.  
- Une fois que ma charge est terminée, je bouge vers le point de sauvetage le plus proche.  

Step 1 :  
- Je scanne les personnes en danger  
- Si je vois une personne en danger, je la sauvegarde.  

Step 2 :  
- J'essaye de communiquer avec un RP si un RP est dans mon rayon de communication.  
   - Si aucun RP n'est dans mon rayon de communication.  
		- J'essaye de voir si je peux envoyer l'information à un drone qui est dans mon network.  
			- Un network est un sous-ensemble de drones qui peuvent communiquer entre eux, ils sont chainées et ils forment un sous-graphe.  
		- Si je ne peux pas, je prends le drone le plus proche dans mon network en terme de distance d'un RP et je lui transfère la resposnabilité de sauver les personnes.  

Step 3 :  
- Je bouge vers le rescue point si je suis le drone le plus proche.  