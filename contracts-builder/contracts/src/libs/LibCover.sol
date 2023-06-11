// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

import {CardTypes} from "../codegen/Types.sol";
import {CoverPosition} from "../codegen/tables/CoverPosition.sol";
import {CurrentHp} from "../codegen/tables/CurrentHp.sol";
import {Position} from "../codegen/tables/Position.sol";

library LibCover {
    function getCoverCard(bytes32 gameKey, bytes32 playerKey, uint32 newX, uint32 newY)
        internal
        view
        returns (bytes32)
    {
        (bytes32 card, bytes32 player, bytes32 card2, bytes32 player2) = CoverPosition.get(gameKey);
        if (player != bytes32(0) && player != playerKey) {
            (bool placed,, uint32 x, uint32 y) = Position.get(card);
            require(placed == true, "cover card was not summoned");
            require(x != 99, "the card is dead");
            // Check max distance using movement speed
            uint32 deltaX = newX > x ? newX - x : x - newX;
            uint32 deltaY = newY > y ? newY - y : y - newY;
            if (deltaX + deltaY <= 2) {
                return card;
            }
            return bytes32(0);
        } else if (player2 != bytes32(0) && player2 == playerKey) {
            (bool placed,, uint32 x, uint32 y) = Position.get(card2);
            require(placed == true, "cover card was not summoned");
            require(x != 99, "the card is dead");
            // Check max distance using movement speed
            uint32 deltaX = newX > x ? newX - x : x - newX;
            uint32 deltaY = newY > y ? newY - y : y - newY;
            if (deltaX + deltaY <= 2) {
                return card2;
            }
            return bytes32(0);
        }
        return bytes32(0);
    }
}
