#!/usr/bin/env python3
"""
Automated Hit Detection System Test
Tests the complete flow: Connect → Receive monsters → Attack → Receive combat event
"""

import asyncio
import websockets
import json
import sys
from datetime import datetime

class HitDetectionTest:
    def __init__(self):
        self.ws = None
        self.player_id = None
        self.monsters = []
        self.tests_passed = {
            'connected': False,
            'received_welcome': False,
            'received_monsters': False,
            'attack_sent': False,
            'combat_event_received': False
        }

    def log(self, message, type='INFO'):
        colors = {
            'INFO': '\033[94m',
            'SUCCESS': '\033[92m',
            'ERROR': '\033[91m',
            'WARN': '\033[93m'
        }
        reset = '\033[0m'
        timestamp = datetime.now().strftime('%H:%M:%S')
        print(f"{colors.get(type, '')}{type:<8}{reset} [{timestamp}] {message}")

    async def connect(self):
        """Connect to WebSocket server"""
        try:
            self.ws = await websockets.connect('ws://localhost:8080/ws')
            self.tests_passed['connected'] = True
            self.log("✅ Connected to server", 'SUCCESS')
            return True
        except Exception as e:
            self.log(f"❌ Failed to connect: {e}", 'ERROR')
            return False

    async def handle_message(self, message):
        """Process incoming WebSocket message"""
        try:
            data = json.loads(message)
            msg_type = data.get('type')

            if msg_type == 'welcome':
                self.player_id = data['data']['player_id']
                self.tests_passed['received_welcome'] = True
                self.log(f"✅ Received welcome, Player ID: {self.player_id}", 'SUCCESS')

            elif msg_type == 'world_state':
                self.monsters = data['data'].get('monsters', [])
                if self.monsters and not self.tests_passed['received_monsters']:
                    self.tests_passed['received_monsters'] = True
                    self.log(f"✅ Received {len(self.monsters)} monsters", 'SUCCESS')
                    for monster in self.monsters:
                        self.log(f"   🧟 {monster['id']} ({monster['name']}) - HP: {monster['health']}/{monster['max_health']}", 'INFO')

            elif msg_type == 'combat_event':
                self.handle_combat_event(data['data'])

            elif msg_type == 'combat_events_batch':
                for event in data['data']['events']:
                    self.handle_combat_event(event)

        except Exception as e:
            self.log(f"Error handling message: {e}", 'ERROR')

    def handle_combat_event(self, event):
        """Process combat event"""
        event_types = ['DAMAGE', 'HEAL', 'BUFF', 'DEBUFF', 'DEATH']
        event_type = event_types[event.get('type', 0)]

        attacker = event.get('attacker_id', 'unknown')
        target = event.get('target_id', 'unknown')
        damage = event.get('damage', 0)
        is_crit = event.get('is_critical', False)
        target_health = event.get('target_health', 0)

        crit_marker = " 💥 CRIT!" if is_crit else ""
        self.log(
            f"⚔️ {event_type}: {attacker} → {target} "
            f"({damage} damage, HP: {target_health}){crit_marker}",
            'SUCCESS' if event_type == 'DAMAGE' else 'INFO'
        )

        if event.get('type') == 0:  # Damage event
            self.tests_passed['combat_event_received'] = True

        if event.get('type') == 4:  # Death event
            self.log(f"💀 {target} has died!", 'SUCCESS')

    async def attack_monster(self, monster_id):
        """Send attack command to server"""
        message = {
            'type': 'attack',
            'player_id': self.player_id,
            'data': {
                'target_id': monster_id,
                'ability_id': 'basic_attack'
            }
        }

        await self.ws.send(json.dumps(message))
        self.tests_passed['attack_sent'] = True
        self.log(f"→ Attack sent to: {monster_id}", 'INFO')

    def print_results(self):
        """Print test results summary"""
        print("\n" + "="*60)
        print("TEST RESULTS")
        print("="*60)

        for test, passed in self.tests_passed.items():
            status = "✅ PASS" if passed else "❌ FAIL"
            print(f"{status} - {test.replace('_', ' ').title()}")

        all_passed = all(self.tests_passed.values())
        print("="*60)
        if all_passed:
            print("🎉 ALL TESTS PASSED! Hit detection system working correctly!")
        else:
            print("⚠️  Some tests failed. Check the logs above.")
        print("="*60 + "\n")

        return all_passed

    async def run_tests(self):
        """Run complete test suite"""
        self.log("🚀 Starting Hit Detection System Test", 'INFO')
        self.log("="*50, 'INFO')

        # Test 1: Connect
        if not await self.connect():
            self.print_results()
            return False

        try:
            # Test 2-3: Wait for welcome and world state
            self.log("⏳ Waiting for welcome message and monsters...", 'INFO')

            timeout = 5
            start_time = asyncio.get_event_loop().time()

            while asyncio.get_event_loop().time() - start_time < timeout:
                try:
                    message = await asyncio.wait_for(self.ws.recv(), timeout=1.0)
                    await self.handle_message(message)

                    # Break once we have monsters
                    if self.monsters:
                        break
                except asyncio.TimeoutError:
                    continue

            if not self.monsters:
                self.log("❌ No monsters received within timeout", 'ERROR')
                self.print_results()
                return False

            # Test 4: Attack first monster
            self.log(f"\n⚔️ Testing attack on {self.monsters[0]['id']}...", 'INFO')
            await self.attack_monster(self.monsters[0]['id'])

            # Test 5: Wait for combat event
            self.log("⏳ Waiting for combat event...", 'INFO')

            timeout = 3
            start_time = asyncio.get_event_loop().time()

            while asyncio.get_event_loop().time() - start_time < timeout:
                try:
                    message = await asyncio.wait_for(self.ws.recv(), timeout=1.0)
                    await self.handle_message(message)

                    # Break once we receive combat event
                    if self.tests_passed['combat_event_received']:
                        break
                except asyncio.TimeoutError:
                    continue

            if not self.tests_passed['combat_event_received']:
                self.log("⚠️ No combat event received within timeout", 'WARN')

            # Optional: Test attacking all monsters
            if len(self.monsters) > 1:
                self.log(f"\n🔥 Bonus test: Attacking all {len(self.monsters)} monsters...", 'INFO')
                for monster in self.monsters[1:]:
                    await self.attack_monster(monster['id'])
                    await asyncio.sleep(0.2)

                # Wait for events
                for _ in range(5):
                    try:
                        message = await asyncio.wait_for(self.ws.recv(), timeout=0.5)
                        await self.handle_message(message)
                    except asyncio.TimeoutError:
                        break

        except Exception as e:
            self.log(f"❌ Test error: {e}", 'ERROR')
        finally:
            await self.ws.close()

        # Print final results
        return self.print_results()

async def main():
    """Main entry point"""
    test = HitDetectionTest()
    success = await test.run_tests()
    sys.exit(0 if success else 1)

if __name__ == '__main__':
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        print("\n\nTest interrupted by user")
        sys.exit(1)
