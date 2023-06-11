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

struct CardToAttack {
    bytes32 card;
    uint32 x;
    uint32 y;
}

contract WhirlwindAxeSystem is System {
    function validate(bytes32 cardKey, bytes32 gameKey, bytes32 playerKey) private view {
        require(Card.get(cardKey), "card does not exist");
        require(ActionReady.get(cardKey) == true, "card already attacked");
        require(OwnedBy.get(cardKey) == playerKey, "the sender is not the owner of the card");
        require(CurrentPlayer.get(gameKey) == playerKey, "it is not the player turn");
        require(AbilityType.get(cardKey) == AbilityTypes.WhirlwindAxe, "card does not have the ability");
    }

    function getDirections(bytes32 cardKey, bytes32 gameKeyGenerated)
        internal
        view
        returns (CardToAttack[8] memory cards)
    {
        // Is the card in the board?
        (bool placed,, uint32 x, uint32 y) = Position.get(cardKey);
        require(placed == true, "card was not summoned");
        CardToAttack[8] memory ret = [
            CardToAttack(bytes32(0), uint32(0), uint32(0)),
            CardToAttack(bytes32(0), uint32(0), uint32(0)),
            CardToAttack(bytes32(0), uint32(0), uint32(0)),
            CardToAttack(bytes32(0), uint32(0), uint32(0)),
            CardToAttack(bytes32(0), uint32(0), uint32(0)),
            CardToAttack(bytes32(0), uint32(0), uint32(0)),
            CardToAttack(bytes32(0), uint32(0), uint32(0)),
            CardToAttack(bytes32(0), uint32(0), uint32(0))
        ];
        bytes32[] memory temp;
        temp = getKeysWithValue(PositionTableId, Position.encode(true, gameKeyGenerated, x - 1, y));
        if (temp.length > 0) {
            ret[0] = CardToAttack(temp[0], x - 1, y);
        }
        temp = getKeysWithValue(PositionTableId, Position.encode(true, gameKeyGenerated, x + 1, y));
        if (temp.length > 0) {
            ret[1] = CardToAttack(temp[0], x + 1, y);
        }
        temp = getKeysWithValue(PositionTableId, Position.encode(true, gameKeyGenerated, x, y - 1));
        if (temp.length > 0) {
            ret[2] = CardToAttack(temp[0], x, y - 1);
        }
        temp = getKeysWithValue(PositionTableId, Position.encode(true, gameKeyGenerated, x, y + 1));
        if (temp.length > 0) {
            ret[3] = CardToAttack(temp[0], x, y + 1);
        }
        temp = getKeysWithValue(PositionTableId, Position.encode(true, gameKeyGenerated, x + 1, y + 1));
        if (temp.length > 0) {
            ret[4] = CardToAttack(temp[0], x + 1, y + 1);
        }
        temp = getKeysWithValue(PositionTableId, Position.encode(true, gameKeyGenerated, x - 1, y + 1));
        if (temp.length > 0) {
            ret[5] = CardToAttack(temp[0], x - 1, y + 1);
        }
        temp = getKeysWithValue(PositionTableId, Position.encode(true, gameKeyGenerated, x + 1, y - 1));
        if (temp.length > 0) {
            ret[6] = CardToAttack(temp[0], x + 1, y - 1);
        }
        temp = getKeysWithValue(PositionTableId, Position.encode(true, gameKeyGenerated, x - 1, y - 1));
        if (temp.length > 0) {
            ret[7] = CardToAttack(temp[0], x - 1, y - 1);
        }
        return ret;
    }

    function limits(bytes32 cardKey, bytes32 gameKeyGenerated, bytes32 playerKey)
        private
        view
        returns (PlacedCardsData memory)
    {
        // Is the card the base
        require(UnitType.get(cardKey) != CardTypes.Base, "can not place move the base");

        // Mana
        require(CurrentMana.get(gameKeyGenerated) >= 3, "no enough mana");
    }

    function whirlwindaxe(bytes32 cardKey) public {
        bytes32 gameKeyGenerated = UsedIn.get(cardKey);
        bytes32 playerKey = addressToEntityKey(_msgSender());
        require(gameKeyGenerated != 0, "game id is incorrect");
        validate(cardKey, gameKeyGenerated, playerKey);

        // limits
        limits(cardKey, gameKeyGenerated, playerKey);

        uint32 attackDmg = 3;

        CardToAttack[8] memory cards = getDirections(cardKey, gameKeyGenerated);
        uint256 i = 0;
        bool baseAlreadyAttacked = false;
        for (i = 0; i < 8; i++) {
            if (cards[i].card != 0) {
                // TODO: validate that the card is not owned by the player
                bytes32 attackedKey = cards[i].card;
                // Check if it's part of the based
                bytes32 isBase = IsBase.get(attackedKey);
                if (isBase != 0) {
                    if (baseAlreadyAttacked) {
                        continue;
                    }
                    attackedKey = isBase;
                    baseAlreadyAttacked = true;
                }

                bytes32 cover = LibCover.getCoverCard(gameKeyGenerated, playerKey, cards[i].x, cards[i].y);
                if (cover != bytes32(0)) {
                    attackedKey = cover;
                }

                uint32 hp = CurrentHp.get(attackedKey);
                if (hp <= attackDmg) {
                    // DEAD
                    CurrentHp.set(attackedKey, 0);
                    Position.set(attackedKey, true, gameKeyGenerated, 99, 99);
                    if (isBase != 0) {
                        // TODO: delete everything
                        PlayerOne.deleteRecord(gameKeyGenerated);
                        PlayerTwo.deleteRecord(gameKeyGenerated);
                        Match.deleteRecord(gameKeyGenerated);
                    }
                } else {
                    // Reduce hp
                    CurrentHp.set(attackedKey, hp - attackDmg);
                }
            }
        }

        ActionReady.set(cardKey, false);
        // Update game status
        CurrentMana.set(gameKeyGenerated, CurrentMana.get(gameKeyGenerated) - 3);
    }
}
