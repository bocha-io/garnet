// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

import { IBaseWorld } from "@latticexyz/world/src/interfaces/IBaseWorld.sol";

import { IAttackSystem } from "./IAttackSystem.sol";
import { ICoverSystem } from "./ICoverSystem.sol";
import { ICreateMatchSystem } from "./ICreateMatchSystem.sol";
import { IDrainSwordSystem } from "./IDrainSwordSystem.sol";
import { IEndTurnSystem } from "./IEndTurnSystem.sol";
import { IJoinMatchSystem } from "./IJoinMatchSystem.sol";
import { IMeteorSystem } from "./IMeteorSystem.sol";
import { IMoveCardSystem } from "./IMoveCardSystem.sol";
import { IPiercingShotSystem } from "./IPiercingShotSystem.sol";
import { IPlaceCardSystem } from "./IPlaceCardSystem.sol";
import { IRegisterSystem } from "./IRegisterSystem.sol";
import { ISidestepSystem } from "./ISidestepSystem.sol";
import { ISurrenderSystem } from "./ISurrenderSystem.sol";
import { IWhirlwindAxeSystem } from "./IWhirlwindAxeSystem.sol";

/**
 * The IWorld interface includes all systems dynamically added to the World
 * during the deploy process.
 */
interface IWorld is
  IBaseWorld,
  IAttackSystem,
  ICoverSystem,
  ICreateMatchSystem,
  IDrainSwordSystem,
  IEndTurnSystem,
  IJoinMatchSystem,
  IMeteorSystem,
  IMoveCardSystem,
  IPiercingShotSystem,
  IPlaceCardSystem,
  IRegisterSystem,
  ISidestepSystem,
  ISurrenderSystem,
  IWhirlwindAxeSystem
{

}