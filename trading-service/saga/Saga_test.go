package saga

import (
    "testing"
)

func TestUpdatePhase(t *testing.T) {
    // Create a new state manager for testing
    manager := SagaStateManager{
        states: make(map[string]*OtcSagaState),
    }

    // Test creating a new entry
    uid := "test-uid-1"
    manager.UpdatePhase(uid, PhaseBuyerReserved)
    
    state, exists := manager.states[uid]
    if !exists {
        t.Errorf("Expected state to exist for UID %s, but it doesn't", uid)
    }
    
    if state.Phase != PhaseBuyerReserved {
        t.Errorf("Expected phase %d, got %d", PhaseBuyerReserved, state.Phase)
    }
    
    // Test updating an existing entry
    manager.UpdatePhase(uid, PhaseSellerReceived)
    
    if manager.states[uid].Phase != PhaseSellerReceived {
        t.Errorf("Expected phase %d after update, got %d", PhaseSellerReceived, manager.states[uid].Phase)
    }
}

func TestGetPhase(t *testing.T) {
    // Create a new state manager for testing
    manager := SagaStateManager{
        states: make(map[string]*OtcSagaState),
    }
    
    // Test getting a non-existent phase
    uid := "test-uid-2"
    phase, exists := manager.GetPhase(uid)
    
    if exists {
        t.Errorf("Expected exists to be false for non-existent UID, got true")
    }
    
    if phase != 0 {
        t.Errorf("Expected default phase 0 for non-existent UID, got %d", phase)
    }
    
    // Test getting an existing phase
    manager.states[uid] = &OtcSagaState{UID: uid, Phase: PhaseOwnershipTransferred}
    
    phase, exists = manager.GetPhase(uid)
    
    if !exists {
        t.Errorf("Expected exists to be true for existing UID, got false")
    }
    
    if phase != PhaseOwnershipTransferred {
        t.Errorf("Expected phase %d, got %d", PhaseOwnershipTransferred, phase)
    }
}

func TestRemove(t *testing.T) {
    // Create a new state manager for testing
    manager := SagaStateManager{
        states: make(map[string]*OtcSagaState),
    }
    
    // Set up test data
    uid := "test-uid-3"
    manager.states[uid] = &OtcSagaState{UID: uid, Phase: PhaseVerified}
    
    // Test removing an existing entry
    manager.Remove(uid)
    
    if _, exists := manager.states[uid]; exists {
        t.Errorf("Expected state to be removed for UID %s, but it still exists", uid)
    }
    
    // Test removing a non-existent entry (should not panic)
    nonExistentUID := "non-existent"
    manager.Remove(nonExistentUID) // This should not panic
}

func TestConcurrentOperations(t *testing.T) {
    // This test checks that the mutex prevents race conditions
    manager := SagaStateManager{
        states: make(map[string]*OtcSagaState),
    }
    
    uid := "test-uid-4"
    
    // Run operations concurrently
    done := make(chan bool)
    
    go func() {
        for i := 0; i < 100; i++ {
            manager.UpdatePhase(uid, SagaPhase(i%6))
        }
        done <- true
    }()
    
    go func() {
        for i := 0; i < 100; i++ {
            manager.GetPhase(uid)
        }
        done <- true
    }()
    
    go func() {
        for i := 0; i < 100; i++ {
            if i%10 == 0 {
                manager.Remove(uid)
            }
        }
        done <- true
    }()
    
    // Wait for all goroutines to finish
    <-done
    <-done
    <-done
    
    // If we got here without deadlock or panic, the test passes
}

func TestPhaseConstants(t *testing.T) {
    // Verify the phase constants have expected values
    phases := []SagaPhase{
        PhaseInit,
        PhaseBuyerReserved,
        PhaseSellerReceived,
        PhaseOwnershipRemoved,
        PhaseOwnershipTransferred,
        PhaseVerified,
    }
    
    for i, phase := range phases {
        if int(phase) != i {
            t.Errorf("Expected phase at index %d to have value %d, got %d", i, i, phase)
        }
    }
}