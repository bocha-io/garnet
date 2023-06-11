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
import {AbilityType} from "../codegen/tables/AbilityType.sol";

import {CoverPosition} from "../codegen/tables/CoverPosition.sol";

contract CoverSystem is System {
    function validate(bytes32 cardKey, bytes32 gameKey, bytes32 playerKey) private view {
        require(Card.get(cardKey), "card does not exist");
        require(ActionReady.get(cardKey) == true, "card already attacked");
        require(OwnedBy.get(cardKey) == playerKey, "the sender is not the owner of the card");
        require(CurrentPlayer.get(gameKey) == playerKey, "it is not the player turn");
        require(AbilityType.get(cardKey) == AbilityTypes.Cover, "card does not have the ability");
    }

    function limits(bytes32 cardKey, bytes32 gameKeyGenerated, bytes32 playerKey)
        private
        view
        returns (PlacedCardsData memory)
    {
        // Is the card the base
        require(UnitType.get(cardKey) != CardTypes.Base, "can not place move the base");

        (bool placed,,,) = Position.get(cardKey);
        require(placed == true, "card was not summoned");

        // Mana
        require(CurrentMana.get(gameKeyGenerated) >= 4, "no enough mana");
    }

    function cover(bytes32 cardKey) public {
        bytes32 gameKeyGenerated = UsedIn.get(cardKey);
        bytes32 playerKey = addressToEntityKey(_msgSender());
        require(gameKeyGenerated != 0, "game id is incorrect");
        validate(cardKey, gameKeyGenerated, playerKey);
        // limits
        limits(cardKey, gameKeyGenerated, playerKey);
        // Check that there is no card in that position
        ActionReady.set(cardKey, false);
        // Update game status
        CurrentMana.set(gameKeyGenerated, CurrentMana.get(gameKeyGenerated) - 4);

        // Create cover position
        (bytes32 card, bytes32 player, bytes32 card2, bytes32 player2) = CoverPosition.get(gameKeyGenerated);
        if (player == playerKey) {
            CoverPosition.set(gameKeyGenerated, cardKey, playerKey, card2, player2);
        } else if (player2 == playerKey) {
            CoverPosition.set(gameKeyGenerated, card, player, cardKey, playerKey);
        } else if (player == bytes32(0)) {
            CoverPosition.set(gameKeyGenerated, cardKey, playerKey, card2, player2);
        } else if (player2 == bytes32(0)) {
            CoverPosition.set(gameKeyGenerated, card, player, cardKey, playerKey);
        } else {
            require(false, "invalid player key for this match");
        }
    }
}
