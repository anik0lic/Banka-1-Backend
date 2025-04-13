package com.banka1.user.DTO.banking;

import com.banka1.user.DTO.banking.helper.AccountStatus;
import com.banka1.user.DTO.banking.helper.AccountSubtype;
import com.banka1.user.DTO.banking.helper.AccountType;
import com.banka1.user.DTO.banking.helper.CurrencyType;
import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class CreateAccountDTO {
    @NotNull(message = "ID vlasnika racuna ne može biti prazan")
    private Long ownerID;

    @NotNull(message = "Izaberi valutu za racun")
    private CurrencyType currency;

    @NotNull(message = "Izaberi tip racuna")
    private AccountType type;

    @NotNull(message = "Izaberi podtip racuna")
    private AccountSubtype subtype;

    @NotNull(message = "Izaberi dnevni limit za potrosnju sredstava sa racuna")
    private Double dailyLimit;

    @NotNull(message = "Izaberi mesecni limit za potrosnju sredstava sa racuna")
    private Double monthlyLimit;

    @NotNull(message = "Izaberi status racuna")
    private AccountStatus status;

    @NotNull(message = "Izaberi da li da se kreiraju kartice za racun")
    private Boolean createCard;

    private Double balance;

    private CreateCompanyDTO companyData;

    public CreateAccountDTO(CreateAccountWithoutOwnerIdDTO createAccountWithoutOwnerIdDTO, Long ownerID) {
        setCurrency(createAccountWithoutOwnerIdDTO.getCurrency());
        setStatus(createAccountWithoutOwnerIdDTO.getStatus());
        setType(createAccountWithoutOwnerIdDTO.getType());
        setSubtype(createAccountWithoutOwnerIdDTO.getSubtype());
        setDailyLimit(createAccountWithoutOwnerIdDTO.getDailyLimit());
        setMonthlyLimit(createAccountWithoutOwnerIdDTO.getMonthlyLimit());
        setOwnerID(ownerID);
        setCreateCard(createAccountWithoutOwnerIdDTO.getCreateCard());
        setBalance(createAccountWithoutOwnerIdDTO.getBalance());
        setCompanyData(createAccountWithoutOwnerIdDTO.getCompanyData());
    }
}
