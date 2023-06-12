// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

// Core
import {System} from "@latticexyz/world/src/System.sol";
import {getKeysWithValue} from "@latticexyz/world/src/modules/keyswithvalue/getKeysWithValue.sol";
// Utils
import {addressToEntityKey} from "../addressToEntityKey.sol";
import {CardTypes} from "../codegen/Types.sol";
import {AbilityTypes} from "../codegen/Types.sol";
// Tables
import {Card} from "../codegen/tables/Card.sol";
import {OwnedBy} from "../codegen/tables/OwnedBy.sol";
import {IsBase} from "../codegen/tables/IsBase.sol";
import {UsedIn} from "../codegen/tables/UsedIn.sol";
import {CurrentPlayer} from "../codegen/tables/CurrentPlayer.sol";
import {PlayerOne} from "../codegen/tables/PlayerOne.sol";
import {UnitType} from "../codegen/tables/UnitType.sol";
import {AbilityType} from "../codegen/tables/AbilityType.sol";
import {Position, PositionTableId} from "../codegen/tables/Position.sol";
import {PlacedCards, PlacedCardsData} from "../codegen/tables/PlacedCards.sol";
import {MapConfig, MapConfigData} from "../codegen/tables/MapConfig.sol";
import {CurrentMana} from "../codegen/tables/CurrentMana.sol";
import {CurrentHp} from "../codegen/tables/CurrentHp.sol";
import {AttackDamage} from "../codegen/tables/AttackDamage.sol";
import {ActionReady} from "../codegen/tables/ActionReady.sol";
import {Match} from "../codegen/tables/Match.sol";
import {PlayerOne} from "../codegen/tables/PlayerOne.sol";
import {PlayerTwo} from "../codegen/tables/PlayerTwo.sol";

import {LibCover} from "../libs/LibCover.sol";

contract MeteorSystem is System {
    function validate(bytes32 cardKey, bytes32 gameKey, bytes32 playerKey) private view {
        require(Card.get(cardKey), "card does not exist");
        require(ActionReady.get(cardKey) == true, "card already attacked");
        require(OwnedBy.get(cardKey) == playerKey, "the sender is not the owner of the card");
        require(CurrentPlayer.get(gameKey) == playerKey, "it is not the player turn");
        require(AbilityType.get(cardKey) == AbilityTypes.Meteor, "card does not have the ability");
    }

    function limits(bytes32 cardKey, bytes32 gameKeyGenerated, bytes32 playerKey, uint32 newX, uint32 newY)
        private
        view
        returns (PlacedCardsData memory)
    {
        // Is the card the base
        require(UnitType.get(cardKey) != CardTypes.Base, "can not place move the base");

        // Is the card in the board?
        (bool placed,, uint32 x, uint32 y) = Position.get(cardKey);
        require(placed == true, "card was not summoned");

        // Check max distance using movement speed
        uint32 deltaX = newX > x ? newX - x : x - newX;
        uint32 deltaY = newY > y ? newY - y : y - newY;
        require(deltaX + deltaY <= 5, "the attacket card too far away");

        // Mana
        require(CurrentMana.get(gameKeyGenerated) >= 4, "no enough mana");
    }

    function meteor(bytes32 cardKey, uint32 newX, uint32 newY) public {
        bytes32 gameKeyGenerated = UsedIn.get(cardKey);
        bytes32 playerKey = addressToEntityKey(_msgSender());
        require(gameKeyGenerated != 0, "game id is incorrect");
        validate(cardKey, gameKeyGenerated, playerKey);

        // limits
        limits(cardKey, gameKeyGenerated, playerKey, newX, newY);

        // Check that there is no card in that position
        bytes32[] memory keysAtPos =
            getKeysWithValue(PositionTableId, Position.encode(true, gameKeyGenerated, newX, newY));
        require(keysAtPos.length > 0, "there is no unit in that position");

        bytes32 attackedKey = keysAtPos[0];

        // Check if it's part of the based
        bytes32 isBase = IsBase.get(attackedKey);
        if (isBase != 0) {
            attackedKey = isBase;
        }

        require(OwnedBy.get(attackedKey) != playerKey, "friendy fire is not enabled");

        bytes32 cover = LibCover.getCoverCard(gameKeyGenerated, playerKey, newX, newY);
        if (cover != bytes32(0)) {
            attackedKey = cover;
        }

        uint32 hp = CurrentHp.get(attackedKey);
        require(keysAtPos.length > 0, "there is no unit in that position");
        uint32 attackDmg = 4;

        if (hp <= attackDmg) {
            // DEAD
            CurrentHp.set(attackedKey, 0);
            Position.set(attackedKey, true, gameKeyGenerated, 99, 99);
            if (attackedKey != cover && isBase != 0) {
                // TODO: delete everything
                PlayerOne.deleteRecord(gameKeyGenerated);
                PlayerTwo.deleteRecord(gameKeyGenerated);
                Match.deleteRecord(gameKeyGenerated);
            }
        } else {
            // Reduce hp
            CurrentHp.set(attackedKey, hp - attackDmg);
        }

        ActionReady.set(cardKey, false);
        // Update game status
        CurrentMana.set(gameKeyGenerated, CurrentMana.get(gameKeyGenerated) - 4);
    }
}
